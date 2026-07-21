package siftly

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

// --- Wire format ---

const (
	legacySnapshotVersion = 1
	snapshotVersion       = 2
	metaVersion           = 1
	rowIDAlgorithm        = "fnv1a-normalized-v1"
)

type rowDTO struct {
	Cols          []string `json:"cols"`
	Height        int      `json:"height,omitempty"` // legacy field; ignored at runtime
	ID            uint64   `json:"id"`
	OriginalIndex int      `json:"originalIndex"`
}

type timeWindowDTO struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

type compactColumnLayout struct {
	Role     ui.ColumnRole
	Visible  bool
	MinWidth int
	Weight   float64
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
	return writeAtomicFile(path, 0o600, func(w io.Writer) error {
		buffer := bufio.NewWriterSize(w, 1<<20)
		if err := writeCompactSnapshot(buffer, m); err != nil {
			return err
		}
		return buffer.Flush()
	})
}

// LoadModel replaces the contents of m with the snapshot from path.
func LoadModel(m *Model, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return loadModelFromReader(m, f)
}

func writeCompactSnapshot(w io.Writer, m *Model) error {
	if err := writeLiteral(w, "{\n\"version\":"); err != nil {
		return err
	}
	if err := writeJSONValue(w, snapshotVersion); err != nil {
		return fmt.Errorf("encode version: %w", err)
	}
	if err := writeLiteral(w, ",\n\"rowID\":"); err != nil {
		return err
	}
	if err := writeJSONValue(w, rowIDAlgorithm); err != nil {
		return fmt.Errorf("encode rowID: %w", err)
	}
	if err := writeLiteral(w, ",\n\"header\":"); err != nil {
		return err
	}
	if err := writeJSONValue(w, snapshotHeaderNames(m.table.header)); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}
	if layouts := snapshotColumnLayouts(m.table.header); len(layouts) > 0 {
		if err := writeLiteral(w, ",\n\"columnLayout\":"); err != nil {
			return err
		}
		if err := writeJSONValue(w, layouts); err != nil {
			return fmt.Errorf("encode column layout: %w", err)
		}
	}
	if err := writeLiteral(w, ",\n\"rows\":[\n"); err != nil {
		return err
	}
	for i, r := range m.table.rows {
		if i > 0 {
			if err := writeLiteral(w, ",\n"); err != nil {
				return err
			}
		}
		if err := writeJSONValue(w, r.Cols); err != nil {
			return fmt.Errorf("encode row %d: %w", i, err)
		}
	}
	if err := writeLiteral(w, "\n]"); err != nil {
		return err
	}

	if spans := buildOriginalIndexSpans(m.table.rows); len(spans) > 0 {
		if err := writeLiteral(w, ",\n\"originalIndexSpans\":"); err != nil {
			return err
		}
		if err := writeJSONValue(w, spans); err != nil {
			return fmt.Errorf("encode original indexes: %w", err)
		}
	}

	if marked := u64KeyToStringMarkMap(m.table.markedRows); len(marked) > 0 {
		if err := writeLiteral(w, ",\n\"marked\":"); err != nil {
			return err
		}
		if err := writeJSONValue(w, marked); err != nil {
			return fmt.Errorf("encode marked rows: %w", err)
		}
	}

	if comments := u64KeyToStringStringMap(m.table.commentRows); len(comments) > 0 {
		if err := writeLiteral(w, ",\n\"comments\":"); err != nil {
			return err
		}
		if err := writeJSONValue(w, comments); err != nil {
			return fmt.Errorf("encode comments: %w", err)
		}
	}

	if timeWin := snapshotTimeWindow(m); timeWin != nil {
		if err := writeLiteral(w, ",\n\"timeWindow\":"); err != nil {
			return err
		}
		if err := writeJSONValue(w, timeWin); err != nil {
			return fmt.Errorf("encode time window: %w", err)
		}
	}

	return writeLiteral(w, "\n}\n")
}

func writeLiteral(w io.Writer, s string) error {
	_, err := io.WriteString(w, s)
	return err
}

func writeJSONValue(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func snapshotHeaderNames(header []ui.ColumnMeta) []string {
	if len(header) == 0 {
		return nil
	}
	names := make([]string, len(header))
	for i := range header {
		names[i] = header[i].Name
	}
	return names
}

func snapshotColumnLayouts(header []ui.ColumnMeta) []compactColumnLayout {
	if len(header) == 0 {
		return nil
	}
	layouts := make([]compactColumnLayout, len(header))
	for i, col := range header {
		layouts[i] = compactColumnLayout{
			Role:     col.Role,
			Visible:  col.Visible,
			MinWidth: col.MinWidth,
			Weight:   col.Weight,
		}
	}
	return layouts
}

func (l compactColumnLayout) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{int(l.Role), l.Visible, l.MinWidth, l.Weight})
}

func (l *compactColumnLayout) UnmarshalJSON(data []byte) error {
	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	if len(values) != 4 {
		return fmt.Errorf("column layout has %d values, want 4", len(values))
	}
	var role int
	if err := json.Unmarshal(values[0], &role); err != nil {
		return fmt.Errorf("role: %w", err)
	}
	var visible bool
	if err := json.Unmarshal(values[1], &visible); err != nil {
		return fmt.Errorf("visible: %w", err)
	}
	var minWidth int
	if err := json.Unmarshal(values[2], &minWidth); err != nil {
		return fmt.Errorf("min width: %w", err)
	}
	var weight float64
	if err := json.Unmarshal(values[3], &weight); err != nil {
		return fmt.Errorf("weight: %w", err)
	}
	*l = compactColumnLayout{
		Role:     ui.ColumnRole(role),
		Visible:  visible,
		MinWidth: minWidth,
		Weight:   weight,
	}
	return nil
}

func snapshotTimeWindow(m *Model) *timeWindowDTO {
	if !m.table.timeWindow.Enabled {
		return nil
	}
	return &timeWindowDTO{
		Enabled: true,
		Start:   m.table.timeWindow.Start.Format(time.RFC3339Nano),
		End:     m.table.timeWindow.End.Format(time.RFC3339Nano),
	}
}

func buildOriginalIndexSpans(rows []Row) [][]int {
	var spans [][]int
	currentDelta := 0
	for i, row := range rows {
		originalIndex := row.OriginalIndex
		if originalIndex <= 0 {
			originalIndex = i + 1
		}
		delta := originalIndex - (i + 1)
		if i == 0 {
			currentDelta = delta
			if delta != 0 {
				spans = append(spans, []int{0, delta})
			}
			continue
		}
		if delta != currentDelta {
			spans = append(spans, []int{i, delta})
			currentDelta = delta
		}
	}
	return spans
}

func loadModelFromReader(m *Model, r io.Reader) error {
	dec := json.NewDecoder(bufio.NewReaderSize(r, 1<<20))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("snapshot: expected JSON object")
	}

	version := legacySnapshotVersion
	var header []ui.ColumnMeta
	var columnLayouts []compactColumnLayout
	var rows []Row
	var marked map[string]string
	var comments map[string]string
	var timeWin *timeWindowDTO
	var originalIndexSpans [][]int

	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		field, ok := tok.(string)
		if !ok {
			return fmt.Errorf("snapshot: expected field name")
		}

		switch field {
		case "version":
			if err := dec.Decode(&version); err != nil {
				return fmt.Errorf("decode version: %w", err)
			}
			if version != legacySnapshotVersion && version != snapshotVersion {
				return fmt.Errorf("snapshot version %d not supported (want %d or %d)", version, legacySnapshotVersion, snapshotVersion)
			}
		case "rowID":
			var algorithm string
			if err := dec.Decode(&algorithm); err != nil {
				return fmt.Errorf("decode rowID: %w", err)
			}
			if algorithm != "" && algorithm != rowIDAlgorithm {
				return fmt.Errorf("snapshot rowID algorithm %q not supported (want %q)", algorithm, rowIDAlgorithm)
			}
		case "header":
			header, err = decodeSnapshotHeader(dec)
			if err != nil {
				return fmt.Errorf("decode header: %w", err)
			}
		case "columnLayout":
			if err := dec.Decode(&columnLayouts); err != nil {
				return fmt.Errorf("decode column layout: %w", err)
			}
		case "rows":
			rows, err = decodeSnapshotRows(dec, version)
			if err != nil {
				return err
			}
		case "originalIndexSpans":
			if err := dec.Decode(&originalIndexSpans); err != nil {
				return fmt.Errorf("decode original indexes: %w", err)
			}
		case "marked":
			if err := dec.Decode(&marked); err != nil {
				return fmt.Errorf("decode marked rows: %w", err)
			}
		case "comments":
			if err := dec.Decode(&comments); err != nil {
				return fmt.Errorf("decode comments: %w", err)
			}
		case "timeWindow":
			if err := dec.Decode(&timeWin); err != nil {
				return fmt.Errorf("decode time window: %w", err)
			}
		default:
			if err := skipJSONValue(dec); err != nil {
				return fmt.Errorf("decode %s: %w", field, err)
			}
		}
	}

	tok, err = dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("snapshot: expected end of JSON object")
	}

	if len(columnLayouts) > 0 {
		if err := applyColumnLayouts(header, columnLayouts); err != nil {
			return err
		}
	}
	if version == snapshotVersion {
		if err := applyOriginalIndexSpans(rows, originalIndexSpans); err != nil {
			return err
		}
	}
	return restoreLoadedSnapshot(m, header, rows, marked, comments, timeWin)
}

func decodeSnapshotHeader(dec *json.Decoder) ([]ui.ColumnMeta, error) {
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}

	var legacy []ui.ColumnMeta
	if err := json.Unmarshal(raw, &legacy); err == nil {
		return legacy, nil
	}

	var names []string
	if err := json.Unmarshal(raw, &names); err != nil {
		return nil, err
	}
	header := make([]ui.ColumnMeta, len(names))
	for i, name := range names {
		header[i] = ui.ColumnMeta{
			Name:     name,
			Index:    i,
			Role:     ui.RoleNormal,
			Visible:  true,
			MinWidth: 8,
			Weight:   1,
		}
	}
	return header, nil
}

func applyColumnLayouts(header []ui.ColumnMeta, layouts []compactColumnLayout) error {
	if len(layouts) != len(header) {
		return fmt.Errorf("column layout count %d does not match header count %d", len(layouts), len(header))
	}
	for i, layout := range layouts {
		header[i].Index = i
		header[i].Role = layout.Role
		header[i].Visible = layout.Visible
		header[i].MinWidth = layout.MinWidth
		header[i].Weight = layout.Weight
		header[i].Width = 0
	}
	return nil
}

func decodeSnapshotRows(dec *json.Decoder, version int) ([]Row, error) {
	if version == legacySnapshotVersion {
		return decodeLegacyRows(dec)
	}
	return decodeCompactRows(dec)
}

func decodeLegacyRows(dec *json.Decoder) ([]Row, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("snapshot rows: expected array")
	}

	rows := make([]Row, 0, 1024)
	rowIndex := 0
	for dec.More() {
		var dto rowDTO
		if err := dec.Decode(&dto); err != nil {
			return nil, fmt.Errorf("decode row %d: %w", rowIndex, err)
		}
		row := fromDTORow(dto)
		if row.OriginalIndex <= 0 {
			row.OriginalIndex = rowIndex + 1
		}
		rows = append(rows, row)
		rowIndex++
	}

	tok, err = dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != ']' {
		return nil, fmt.Errorf("snapshot rows: expected end of array")
	}
	return rows, nil
}

func decodeCompactRows(dec *json.Decoder) ([]Row, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("snapshot rows: expected array")
	}

	rows := make([]Row, 0, 1024)
	rowIndex := 0
	for dec.More() {
		var cols []string
		if err := dec.Decode(&cols); err != nil {
			return nil, fmt.Errorf("decode row %d: %w", rowIndex, err)
		}
		row := Row{
			Cols:          cols,
			OriginalIndex: rowIndex + 1,
		}
		row.ID = row.ComputeID()
		rows = append(rows, row)
		rowIndex++
	}

	tok, err = dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != ']' {
		return nil, fmt.Errorf("snapshot rows: expected end of array")
	}
	return rows, nil
}

func applyOriginalIndexSpans(rows []Row, spans [][]int) error {
	if len(spans) == 0 {
		return nil
	}
	prevStart := -1
	for _, span := range spans {
		if len(span) != 2 {
			return fmt.Errorf("invalid originalIndexSpans entry %v", span)
		}
		start := span[0]
		if start < 0 || start >= len(rows) {
			return fmt.Errorf("originalIndexSpans start %d out of range", start)
		}
		if start <= prevStart {
			return fmt.Errorf("originalIndexSpans start %d not greater than previous %d", start, prevStart)
		}
		prevStart = start
	}

	spanIndex := 0
	delta := 0
	for i := range rows {
		if spanIndex < len(spans) && spans[spanIndex][0] == i {
			delta = spans[spanIndex][1]
			spanIndex++
		}
		originalIndex := i + 1 + delta
		if originalIndex <= 0 {
			return fmt.Errorf("original index for row %d is %d", i, originalIndex)
		}
		rows[i].OriginalIndex = originalIndex
	}
	return nil
}

func skipJSONValue(dec *json.Decoder) error {
	var discard json.RawMessage
	return dec.Decode(&discard)
}

func restoreLoadedSnapshot(
	m *Model,
	header []ui.ColumnMeta,
	rows []Row,
	marked map[string]string,
	comments map[string]string,
	timeWin *timeWindowDTO,
) error {
	m.dirty = false
	m.clearRowRangeSelection()
	m.table.header = append([]ui.ColumnMeta(nil), header...)
	m.table.rows = rows
	m.table.showOnlyMarked = false
	m.table.filterPattern = ""
	m.table.filterEnabled = false
	m.table.filterRegex = nil
	m.table.filterWholeRow = false
	m.table.filteredIndices = nil
	m.table.sortEnabled = false
	m.table.sortColumn = -1
	m.table.sortDesc = false
	m.table.rowOrder = nil
	m.table.searchColumns = nil
	m.table.rowTimes = nil
	m.table.rowHasTimes = nil
	m.table.hasTimeBounds = false
	m.table.timeMin = time.Time{}
	m.table.timeMax = time.Time{}
	m.table.timeColumnIndex = -1
	m.table.derivedTimeData = false
	m.table.timeWindow = featuretimewindow.Window{}

	var errMarks, errComments error
	m.table.markedRows, errMarks = parseUintKeyMapMark(marked)
	if errMarks != nil {
		return errMarks
	}
	m.table.commentRows, errComments = parseUintKeyMapString(comments)
	if errComments != nil {
		return errComments
	}

	if timeWin != nil {
		start, err := time.Parse(time.RFC3339Nano, timeWin.Start)
		if err != nil && timeWin.Start != "" {
			return fmt.Errorf("invalid timeWindow start: %w", err)
		}
		end, err := time.Parse(time.RFC3339Nano, timeWin.End)
		if err != nil && timeWin.End != "" {
			return fmt.Errorf("invalid timeWindow end: %w", err)
		}
		m.table.timeWindow = featuretimewindow.Window{
			Enabled: timeWin.Enabled,
			Start:   start,
			End:     end,
		}
	}

	return nil
}

// SaveMeta writes only marks/comments so they can be re-applied after a fresh CSV import.
func SaveMeta(m *Model, path string) error {
	dto := metaOnlyDTO{
		Version:  metaVersion,
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
	if dto.Version != metaVersion {
		return fmt.Errorf("meta version %d not supported (want %d)", dto.Version, metaVersion)
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
