package siftly

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

// --- Wire format ---

const snapshotVersion = 1

type rowDTO struct {
	Cols          []string `json:"cols"`
	Height        int      `json:"height,omitempty"` // legacy field; ignored at runtime
	ID            uint64   `json:"id"`
	OriginalIndex int      `json:"originalIndex"`
}

type snapshotDTO struct {
	Version  int               `json:"version"`
	Header   []ui.ColumnMeta   `json:"header"`
	Rows     []rowDTO          `json:"rows"`
	Marked   map[string]string `json:"marked"`   // MarkColor as string; uint64 keys stringified
	Comments map[string]string `json:"comments"` // uint64 keys stringified
	TimeWin  *timeWindowDTO    `json:"timeWindow,omitempty"`
	Note     string            `json:"note,omitempty"`
}

type timeWindowDTO struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

type metaOnlyDTO struct {
	Version  int               `json:"version"`
	Marked   map[string]string `json:"marked"`
	Comments map[string]string `json:"comments"`
}

// --- Conversions ---

func toDTORow(r Row) rowDTO {
	return rowDTO{
		Cols:          append([]string(nil), r.Cols...),
		ID:            r.ID,
		OriginalIndex: r.OriginalIndex,
	}
}

func fromDTORow(d rowDTO) Row {
	return Row{
		Cols:          append([]string(nil), d.Cols...),
		ID:            d.ID,
		OriginalIndex: d.OriginalIndex,
	}
}

func u64KeyToStringMarkMap(in map[uint64]ui.MarkColor) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[strconv.FormatUint(k, 10)] = string(v)
	}
	return out
}

func u64KeyToStringStringMap(in map[uint64]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[strconv.FormatUint(k, 10)] = v
	}
	return out
}

func parseUintKeyMapMark(in map[string]string) (map[uint64]ui.MarkColor, error) {
	out := make(map[uint64]ui.MarkColor, len(in))
	for ks, vs := range in {
		k, err := strconv.ParseUint(ks, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint64 key %q: %w", ks, err)
		}
		out[k] = sanitizeMarkColor(vs)
	}
	return out, nil
}

func parseUintKeyMapString(in map[string]string) (map[uint64]string, error) {
	out := make(map[uint64]string, len(in))
	for ks, vs := range in {
		k, err := strconv.ParseUint(ks, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint64 key %q: %w", ks, err)
		}
		out[k] = vs
	}
	return out, nil
}

// Accept only known values; anything else becomes MarkNone.
func sanitizeMarkColor(s string) ui.MarkColor {
	switch ui.MarkColor(s) {
	case ui.MarkNone, ui.MarkRed, ui.MarkGreen, ui.MarkAmber:
		return ui.MarkColor(s)
	default:
		return ui.MarkNone
	}
}

// --- Public API ---

// ExportModel writes the *currently filtered* rows to a CSV file,
// including mark color and comment as additional columns.
func ExportModel(m *Model, path string) error {
	// Open file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("open export file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Build header: original columns + Mark + Comment
	header := make([]string, 0, len(m.table.header)+2)
	for _, col := range m.table.header {
		header = append(header, col.Name)
	}
	header = append(header, "Mark", "Comment")

	if err := w.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Decide which indices to export:
	// if filteredIndices is empty, fall back to all rows.
	indices := m.table.filteredIndices
	if len(indices) == 0 {
		indices = make([]int, len(m.table.rows))
		for i := range m.table.rows {
			indices[i] = i
		}
	}

	// Export each visible row
	for _, idx := range indices {
		// sanity check
		if idx < 0 || idx >= len(m.table.rows) {
			return fmt.Errorf("filtered index %d out of range", idx)
		}
		r := m.table.rows[idx]

		// row data: original cols
		out := append([]string(nil), r.Cols...)

		// append mark + comment using the row's id
		mark := ""
		if c, ok := m.table.markedRows[r.ID]; ok {
			mark = string(c)
		}

		comment := ""
		if c, ok := m.table.commentRows[r.ID]; ok {
			comment = c
		}

		out = append(out, mark, comment)

		if err := w.Write(out); err != nil {
			return fmt.Errorf("write row %d: %w", idx, err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}

	return nil
}

// SaveModel writes the entire model to a JSON file.
func SaveModel(m *Model, path string) error {
	dto := snapshotDTO{
		Version:  snapshotVersion,
		Header:   nil, // filled below
		Rows:     make([]rowDTO, 0, len(m.table.rows)),
		Marked:   u64KeyToStringMarkMap(m.table.markedRows),
		Comments: u64KeyToStringStringMap(m.table.commentRows),
	}
	dto.TimeWin = &timeWindowDTO{
		Enabled: m.table.timeWindow.Enabled,
		Start:   m.table.timeWindow.Start.Format(time.RFC3339Nano),
		End:     m.table.timeWindow.End.Format(time.RFC3339Nano),
	}

	// Copy header metadata
	if len(m.table.header) > 0 {
		dto.Header = make([]ui.ColumnMeta, len(m.table.header))
		copy(dto.Header, m.table.header)
	}

	// Copy rows
	for _, r := range m.table.rows {
		dto.Rows = append(dto.Rows, toDTORow(r))
	}

	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// LoadModel replaces the contents of m with the snapshot from path.
func LoadModel(m *Model, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var dto snapshotDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	if dto.Version != snapshotVersion {
		return fmt.Errorf("snapshot version %d not supported (want %d)", dto.Version, snapshotVersion)
	}

	// Restore header
	m.table.header = m.table.header[:0]
	if len(dto.Header) > 0 {
		m.table.header = make([]ui.ColumnMeta, len(dto.Header))
		copy(m.table.header, dto.Header)
	}

	// Restore rows
	m.table.rows = m.table.rows[:0]
	for _, dr := range dto.Rows {
		m.table.rows = append(m.table.rows, fromDTORow(dr))
	}

	// Restore marks/comments
	var errMarks, errComments error
	m.table.markedRows, errMarks = parseUintKeyMapMark(dto.Marked)
	if errMarks != nil {
		return errMarks
	}
	m.table.commentRows, errComments = parseUintKeyMapString(dto.Comments)
	if errComments != nil {
		return errComments
	}

	// Restore time window (bounds recomputed in InitialiseUI)
	if dto.TimeWin != nil {
		start, err := time.Parse(time.RFC3339Nano, dto.TimeWin.Start)
		if err != nil && dto.TimeWin.Start != "" {
			return fmt.Errorf("invalid timeWindow start: %w", err)
		}
		end, err := time.Parse(time.RFC3339Nano, dto.TimeWin.End)
		if err != nil && dto.TimeWin.End != "" {
			return fmt.Errorf("invalid timeWindow end: %w", err)
		}
		m.table.timeWindow = featuretimewindow.Window{
			Enabled: dto.TimeWin.Enabled,
			Start:   start,
			End:     end,
		}
	}

	return nil
}

// SaveMeta writes only marks/comments so they can be re-applied after a fresh CSV import.
func SaveMeta(m *Model, path string) error {
	dto := metaOnlyDTO{
		Version:  snapshotVersion,
		Marked:   u64KeyToStringMarkMap(m.table.markedRows),
		Comments: u64KeyToStringStringMap(m.table.commentRows),
	}
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// LoadMeta merges marks/comments into m, only for rows currently present (by ID).
func LoadMeta(m *Model, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var dto metaOnlyDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	if dto.Version != snapshotVersion {
		return fmt.Errorf("meta version %d not supported (want %d)", dto.Version, snapshotVersion)
	}

	if m.table.markedRows == nil {
		m.table.markedRows = make(map[uint64]ui.MarkColor)
	}
	if m.table.commentRows == nil {
		m.table.commentRows = make(map[uint64]string)
	}

	present := make(map[uint64]struct{}, len(m.table.rows))
	for _, r := range m.table.rows {
		present[r.ID] = struct{}{}
	}

	for ks, vs := range dto.Marked {
		k, err := strconv.ParseUint(ks, 10, 64)
		if err != nil {
			return err
		}
		if _, ok := present[k]; ok {
			m.table.markedRows[k] = sanitizeMarkColor(vs)
		}
	}
	for ks, vs := range dto.Comments {
		k, err := strconv.ParseUint(ks, 10, 64)
		if err != nil {
			return err
		}
		if _, ok := present[k]; ok {
			m.table.commentRows[k] = vs
		}
	}

	return nil
}
