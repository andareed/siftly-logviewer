package siftly

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	recoveryVersion  = 1
	recoveryDebounce = 750 * time.Millisecond
)

var errRecoveryUnavailable = errors.New("recovery unavailable")

type recoveryFlushMsg struct {
	Seq int
}

type recoveryWriteCompleteMsg struct {
	Seq int
	Err error
}

type recoverySourceDTO struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	ModTimeNano int64  `json:"modTimeNano"`
	Rows        int    `json:"rows"`
	FirstRowID  uint64 `json:"firstRowID,omitempty"`
	LastRowID   uint64 `json:"lastRowID,omitempty"`
	HeaderHash  string `json:"headerHash"`
}

type recoveryColumnDTO struct {
	Name      string        `json:"name"`
	Index     int           `json:"index"`
	Role      ui.ColumnRole `json:"role"`
	Visible   bool          `json:"visible"`
	Frozen    bool          `json:"frozen,omitempty"`
	MinWidth  int           `json:"minWidth"`
	Weight    float64       `json:"weight"`
	WrapLines int           `json:"wrapLines,omitempty"`
}

type recoveryStateDTO struct {
	Columns        []recoveryColumnDTO `json:"columns"`
	Marked         map[string]string   `json:"marked,omitempty"`
	Comments       map[string]string   `json:"comments,omitempty"`
	TimeWindow     *timeWindowDTO      `json:"timeWindow,omitempty"`
	SortEnabled    bool                `json:"sortEnabled,omitempty"`
	SortColumn     int                 `json:"sortColumn,omitempty"`
	SortDesc       bool                `json:"sortDesc,omitempty"`
	FilterPattern  string              `json:"filterPattern,omitempty"`
	FilterEnabled  bool                `json:"filterEnabled,omitempty"`
	FilterWholeRow bool                `json:"filterWholeRow,omitempty"`
	ShowOnlyMarked bool                `json:"showOnlyMarked,omitempty"`
	GraphOpen      bool                `json:"graphOpen,omitempty"`
}

type recoverySnapshotDTO struct {
	Version int               `json:"version"`
	SavedAt string            `json:"savedAt"`
	Source  recoverySourceDTO `json:"source"`
	State   recoveryStateDTO  `json:"state"`
}

type pendingRecovery struct {
	state      trackedState
	path       string
	sourceName string
	savedAt    string
	contents   string
}

// SetRecoveryPath overrides the source-adjacent recovery path. It is primarily
// useful to wrappers that manage their own state directory.
func (m *Model) SetRecoveryPath(path string) {
	m.changes.recoveryPathOverride = strings.TrimSpace(path)
}

func (m *Model) scheduleRecovery() {
	m.changes.recoverySeq++
	seq := m.changes.recoverySeq
	if _, err := m.recoveryFilePath(); err != nil {
		m.changes.pendingRecoveryCmd = nil
		return
	}
	m.changes.pendingRecoveryCmd = tea.Tick(recoveryDebounce, func(time.Time) tea.Msg {
		return recoveryFlushMsg{Seq: seq}
	})
}

func (m *Model) takePendingRecoveryCmd() tea.Cmd {
	cmd := m.changes.pendingRecoveryCmd
	m.changes.pendingRecoveryCmd = nil
	return cmd
}

func (m *Model) handleRecoveryFlush(msg recoveryFlushMsg) tea.Cmd {
	if msg.Seq != m.changes.recoverySeq {
		return nil
	}
	path, err := m.recoveryFilePath()
	if err != nil {
		return nil
	}
	if !m.dirty {
		return func() tea.Msg {
			err := removeRecoveryFile(path)
			return recoveryWriteCompleteMsg{Seq: msg.Seq, Err: err}
		}
	}

	snapshot, err := m.buildRecoverySnapshot()
	if err != nil {
		return func() tea.Msg { return recoveryWriteCompleteMsg{Seq: msg.Seq, Err: err} }
	}
	return func() tea.Msg {
		err := writeRecoveryFile(path, snapshot)
		return recoveryWriteCompleteMsg{Seq: msg.Seq, Err: err}
	}
}

func (m *Model) handleRecoveryWriteComplete(msg recoveryWriteCompleteMsg) tea.Cmd {
	if msg.Seq != m.changes.recoverySeq || msg.Err == nil {
		return nil
	}
	logging.Errorf("recovery snapshot: %v", msg.Err)
	return m.view.notice.Start("Recovery snapshot error", "error", noticeDuration)
}

func (m *Model) writeRecoverySnapshotNow() error {
	if !m.dirty {
		return nil
	}
	path, err := m.recoveryFilePath()
	if err != nil {
		if errors.Is(err, errRecoveryUnavailable) {
			return nil
		}
		return err
	}
	snapshot, err := m.buildRecoverySnapshot()
	if err != nil {
		return err
	}
	return writeRecoveryFile(path, snapshot)
}

func (m *Model) prepareRecoverySnapshot() {
	sidecarPath, err := m.recoveryFilePath()
	if err != nil {
		return
	}
	path := sidecarPath
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		legacyPath, legacyErr := m.legacyRecoveryFilePath()
		if legacyErr != nil {
			return
		}
		data, err = os.ReadFile(legacyPath)
		if err == nil {
			path = legacyPath
		}
	}
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		logging.Errorf("read recovery snapshot: %v", err)
		return
	}

	var snapshot recoverySnapshotDTO
	if err := json.Unmarshal(data, &snapshot); err != nil {
		logging.Errorf("decode recovery snapshot: %v", err)
		_ = removeRecoveryFile(path)
		return
	}
	if snapshot.Version != recoveryVersion {
		logging.Infof("ignoring recovery version %d", snapshot.Version)
		_ = removeRecoveryFile(path)
		return
	}
	source, err := m.currentRecoverySource()
	if err != nil || source != snapshot.Source {
		logging.Infof("discarding stale recovery snapshot %q", path)
		_ = removeRecoveryFile(path)
		return
	}

	state, err := trackedStateFromRecoveryDTO(snapshot.State)
	if err != nil {
		logging.Errorf("invalid recovery snapshot state: %v", err)
		_ = removeRecoveryFile(path)
		return
	}
	if !trackedColumnsEqualCurrent(state.Columns, m.table.header) {
		logging.Errorf("recovery columns do not match the current source")
		_ = removeRecoveryFile(path)
		return
	}
	if state.FilterPattern != "" {
		if _, err := regexp.Compile(state.FilterPattern); err != nil {
			logging.Errorf("invalid recovery filter: %v", err)
			_ = removeRecoveryFile(path)
			return
		}
	}

	baseline := m.changes.baseline
	state = normalizeRecoveryColumnWrapping(state, baseline)
	if persistentStatesEqual(state, baseline) {
		_ = removeRecoveryFile(path)
		return
	}
	if path != sidecarPath {
		if err := writeRecoveryFile(sidecarPath, snapshot); err != nil {
			logging.Errorf("migrate legacy recovery snapshot: %v", err)
		} else {
			_ = removeRecoveryFile(path)
			path = sidecarPath
		}
	}
	m.changes.pendingRecovery = &pendingRecovery{
		state:      state,
		path:       path,
		sourceName: filepath.Base(source.Path),
		savedAt:    recoverySavedAtLabel(snapshot.SavedAt),
		contents:   recoveryContentsLabel(state, baseline),
	}
}

func (m *Model) restorePendingRecovery() tea.Cmd {
	pending := m.changes.pendingRecovery
	if pending == nil {
		return m.view.notice.Start("No recovery is available", "warn", noticeDuration)
	}
	m.changes.pendingRecovery = nil

	baseline := m.changes.baseline
	m.restoreTrackedState(pending.state, true)
	restored := m.captureTrackedState()
	if persistentStatesEqual(restored, baseline) {
		_ = removeRecoveryFile(pending.path)
		return m.view.notice.Start("Recovery contained no unsaved changes", "warn", noticeDuration)
	}

	m.pushUndo(historyEntry{state: baseline, label: "recovered changes", size: trackedStateSize(baseline)})
	m.changes.current = restored
	m.updateDirtyFromBaseline()
	return m.view.notice.Start("Unsaved changes restored", "warn", noticeDuration)
}

func (m *Model) discardPendingRecovery() tea.Cmd {
	pending := m.changes.pendingRecovery
	if pending == nil {
		return m.view.notice.Start("No recovery is available", "warn", noticeDuration)
	}
	m.changes.pendingRecovery = nil
	if err := removeRecoveryFile(pending.path); err != nil {
		logging.Errorf("discard recovery snapshot: %v", err)
		return m.view.notice.Start("Could not remove recovery file", "error", noticeDuration)
	}
	return m.view.notice.Start("Recovery discarded", "success", noticeDuration)
}

func (m *Model) buildRecoverySnapshot() (recoverySnapshotDTO, error) {
	source, err := m.currentRecoverySource()
	if err != nil {
		return recoverySnapshotDTO{}, err
	}
	return recoverySnapshotDTO{
		Version: recoveryVersion,
		SavedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Source:  source,
		State:   recoveryDTOFromTrackedState(m.captureTrackedState()),
	}, nil
}

func (m *Model) currentRecoverySource() (recoverySourceDTO, error) {
	documentPath := m.recoveryDocumentPath()
	if documentPath == "" {
		return recoverySourceDTO{}, errRecoveryUnavailable
	}
	absPath, err := filepath.Abs(documentPath)
	if err != nil {
		return recoverySourceDTO{}, err
	}
	absPath = filepath.Clean(absPath)
	info, err := os.Stat(absPath)
	if err != nil {
		return recoverySourceDTO{}, err
	}

	source := recoverySourceDTO{
		Path:        absPath,
		Size:        info.Size(),
		ModTimeNano: info.ModTime().UnixNano(),
		Rows:        len(m.table.rows),
		HeaderHash:  recoveryHeaderHash(m.table.header),
	}
	if len(m.table.rows) > 0 {
		source.FirstRowID = m.table.rows[0].ID
		source.LastRowID = m.table.rows[len(m.table.rows)-1].ID
	}
	return source, nil
}

func (m *Model) recoveryDocumentPath() string {
	if strings.TrimSpace(m.fileName) != "" {
		return m.fileName
	}
	return strings.TrimSpace(m.InitialPath)
}

func (m *Model) recoveryFilePath() (string, error) {
	if m.changes.recoveryPathOverride != "" {
		return m.changes.recoveryPathOverride, nil
	}
	documentPath := m.recoveryDocumentPath()
	if documentPath == "" {
		return "", errRecoveryUnavailable
	}
	absPath, err := filepath.Abs(documentPath)
	if err != nil {
		return "", err
	}
	absPath = filepath.Clean(absPath)
	name := filepath.Base(absPath)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "", errRecoveryUnavailable
	}
	return filepath.Join(
		filepath.Dir(absPath),
		"."+name+".siftly-recovery-"+recoveryOwnerToken()+".json",
	), nil
}

func (m *Model) legacyRecoveryFilePath() (string, error) {
	if m.changes.recoveryPathOverride != "" {
		return "", errRecoveryUnavailable
	}
	documentPath := m.recoveryDocumentPath()
	if documentPath == "" {
		return "", errRecoveryUnavailable
	}
	absPath, err := filepath.Abs(documentPath)
	if err != nil {
		return "", err
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		return "", errRecoveryUnavailable
	}
	digest := sha256.Sum256([]byte(filepath.Clean(absPath)))
	name := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))
	if name == "" || name == "." {
		name = "document"
	}
	return filepath.Join(cacheDir, "siftly", "recovery", name+"-"+hex.EncodeToString(digest[:8])+".json"), nil
}

func recoveryOwnerToken() string {
	current, err := user.Current()
	if err == nil {
		if token := sanitizeRecoveryOwner(current.Uid); token != "" {
			return token
		}
		if token := sanitizeRecoveryOwner(current.Username); token != "" {
			return token
		}
	}
	return "user"
}

func sanitizeRecoveryOwner(value string) string {
	value = strings.TrimSpace(value)
	var token strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-',
			r == '_':
			token.WriteRune(r)
		default:
			token.WriteByte('-')
		}
	}
	result := strings.Trim(token.String(), "-_")
	if result != "" && len(result) <= 32 {
		return result
	}
	if value == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(value))
	return "u-" + hex.EncodeToString(digest[:8])
}

func recoverySavedAtLabel(value string) string {
	savedAt, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return savedAt.Local().Format("02 Jan 2006 15:04:05 MST")
}

func recoveryContentsLabel(state, baseline trackedState) string {
	parts := make([]string, 0, 8)
	if count := len(state.MarkedRows); count > 0 {
		parts = append(parts, pluralCount(count, "mark", "marks"))
	}
	if count := len(state.CommentRows); count > 0 {
		parts = append(parts, pluralCount(count, "comment", "comments"))
	}
	if !reflect.DeepEqual(state.Columns, baseline.Columns) {
		parts = append(parts, "column layout")
	}
	if state.TimeWindow != baseline.TimeWindow {
		parts = append(parts, "time window")
	}
	if state.SortEnabled {
		parts = append(parts, "sorted view")
	}
	if state.FilterPattern != "" {
		parts = append(parts, "filter")
	}
	if state.ShowOnlyMarked {
		parts = append(parts, "marks-only view")
	}
	if state.GraphOpen != baseline.GraphOpen {
		parts = append(parts, "graph state")
	}
	if len(parts) == 0 {
		return "saved session state"
	}
	return strings.Join(parts, ", ")
}

func normalizeRecoveryColumnWrapping(state, baseline trackedState) trackedState {
	wrapLines := make(map[[2]string]int, len(baseline.Columns))
	for _, column := range baseline.Columns {
		key := [2]string{strconv.Itoa(column.Index), column.Name}
		wrapLines[key] = column.WrapLines
	}
	for i := range state.Columns {
		if state.Columns[i].WrapLines != 0 {
			continue
		}
		key := [2]string{strconv.Itoa(state.Columns[i].Index), state.Columns[i].Name}
		state.Columns[i].WrapLines = wrapLines[key]
	}
	return state
}

func pluralCount(count int, singular, plural string) string {
	if count == 1 {
		return "1 " + singular
	}
	return strconv.Itoa(count) + " " + plural
}

func recoveryHeaderHash(header []ui.ColumnMeta) string {
	type sourceColumn struct {
		Index int
		Name  string
	}
	columns := make([]sourceColumn, len(header))
	for i, column := range header {
		columns[i] = sourceColumn{Index: column.Index, Name: column.Name}
	}
	sort.Slice(columns, func(i, j int) bool {
		if columns[i].Index == columns[j].Index {
			return columns[i].Name < columns[j].Name
		}
		return columns[i].Index < columns[j].Index
	})
	h := sha256.New()
	for _, column := range columns {
		_, _ = io.WriteString(h, strconv.Itoa(column.Index))
		_, _ = io.WriteString(h, "\x00")
		_, _ = io.WriteString(h, column.Name)
		_, _ = io.WriteString(h, "\x00")
	}
	return hex.EncodeToString(h.Sum(nil))
}

func recoveryDTOFromTrackedState(state trackedState) recoveryStateDTO {
	columns := make([]recoveryColumnDTO, len(state.Columns))
	for i, column := range state.Columns {
		columns[i] = recoveryColumnDTO(column)
	}
	result := recoveryStateDTO{
		Columns:        columns,
		Marked:         u64KeyToStringMarkMap(state.MarkedRows),
		Comments:       u64KeyToStringStringMap(state.CommentRows),
		SortEnabled:    state.SortEnabled,
		SortColumn:     state.SortColumn,
		SortDesc:       state.SortDesc,
		FilterPattern:  state.FilterPattern,
		FilterEnabled:  state.FilterEnabled,
		FilterWholeRow: state.FilterWholeRow,
		ShowOnlyMarked: state.ShowOnlyMarked,
		GraphOpen:      state.GraphOpen,
	}
	if state.TimeWindow.Enabled {
		result.TimeWindow = &timeWindowDTO{
			Enabled: true,
			Start:   state.TimeWindow.Start.Format(time.RFC3339Nano),
			End:     state.TimeWindow.End.Format(time.RFC3339Nano),
		}
	}
	return result
}

func trackedStateFromRecoveryDTO(dto recoveryStateDTO) (trackedState, error) {
	columns := make([]trackedColumn, len(dto.Columns))
	for i, column := range dto.Columns {
		columns[i] = trackedColumn(column)
	}
	marked, err := parseUintKeyMapMark(dto.Marked)
	if err != nil {
		return trackedState{}, err
	}
	comments, err := parseUintKeyMapString(dto.Comments)
	if err != nil {
		return trackedState{}, err
	}
	state := trackedState{
		Columns:        columns,
		MarkedRows:     marked,
		CommentRows:    comments,
		SortEnabled:    dto.SortEnabled,
		SortColumn:     dto.SortColumn,
		SortDesc:       dto.SortDesc,
		FilterPattern:  dto.FilterPattern,
		FilterEnabled:  dto.FilterEnabled,
		FilterWholeRow: dto.FilterWholeRow,
		ShowOnlyMarked: dto.ShowOnlyMarked,
		GraphOpen:      dto.GraphOpen,
	}
	if dto.TimeWindow != nil {
		start, err := time.Parse(time.RFC3339Nano, dto.TimeWindow.Start)
		if err != nil {
			return trackedState{}, err
		}
		end, err := time.Parse(time.RFC3339Nano, dto.TimeWindow.End)
		if err != nil {
			return trackedState{}, err
		}
		state.TimeWindow = featuretimewindow.Window{Enabled: dto.TimeWindow.Enabled, Start: start, End: end}
	}
	return state, nil
}

func writeRecoveryFile(path string, snapshot recoverySnapshotDTO) error {
	return writeAtomicFile(path, 0o600, func(w io.Writer) error {
		buffer := bufio.NewWriterSize(w, 64<<10)
		encoder := json.NewEncoder(buffer)
		if err := encoder.Encode(snapshot); err != nil {
			return err
		}
		return buffer.Flush()
	})
}

func removeRecoveryFile(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (m *Model) removeRecoverySnapshotNow() error {
	path, err := m.recoveryFilePath()
	if err != nil {
		if errors.Is(err, errRecoveryUnavailable) {
			return nil
		}
		return err
	}
	return removeRecoveryFile(path)
}

func writeAtomicFile(path string, mode fs.FileMode, write func(io.Writer) error) (err error) {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".siftly-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		_ = temp.Close()
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()
	if err = temp.Chmod(mode); err != nil {
		return err
	}
	if err = write(temp); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
