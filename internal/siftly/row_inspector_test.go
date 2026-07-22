package siftly

import (
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

func TestRowInspectorIncludesCompleteRowMetadataAndReorderedFields(t *testing.T) {
	const longMessage = "fstool_va request continued onto another display line with sequence 48291"
	row := Row{
		ID:            42,
		Cols:          []string{"2026-07-21T09:30:00Z", "nascent", longMessage},
		OriginalIndex: 117,
	}
	m := Model{
		cursor: 0,
		table: tableState{
			header: []ui.ColumnMeta{
				{Name: "message", Index: 2},
				{Name: "timestamp", Index: 0},
				{Name: "host", Index: 1, Visible: false},
			},
			rows:            []Row{row},
			filteredIndices: []int{0},
			markedRows:      map[uint64]ui.MarkColor{42: ui.MarkGreen},
			commentRows:     map[uint64]string{42: "Investigate this row before handover"},
		},
		view: viewState{inspector: rowInspectorState{selectedField: 0}},
	}

	content, _ := m.buildInspectorContent(row, 0, 34)
	content = stripANSI(content)
	normalizedContent := strings.Join(strings.Fields(content), " ")
	for _, want := range []string{
		"Investigate this row before handover",
		"message",
		longMessage,
		"timestamp",
		"2026-07-21T09:30:00Z",
		"host",
		"nascent",
	} {
		if !strings.Contains(normalizedContent, strings.Join(strings.Fields(want), " ")) {
			t.Fatalf("inspector content missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(content, "Original row") {
		t.Fatalf("source row metadata should not be repeated in the inspector body:\n%s", content)
	}

	m.inspectorPort = viewport.New(92, 4)
	m.inspectorPort.SetContent(content)
	topBorder := strings.Split(stripANSI(m.rowInspectorView(96)), "\n")[0]
	for _, want := range []string{"Details", "Source 117", "GREEN"} {
		if !strings.Contains(topBorder, want) {
			t.Fatalf("inspector header missing %q: %q", want, topBorder)
		}
	}

	name, value, ok := m.inspectorFieldClipboardText()
	if !ok || name != "message" || value != longMessage {
		t.Fatalf("unexpected selected field: name=%q value=%q ok=%t", name, value, ok)
	}
}

func TestDenseTableSummaryDoesNotAlterInspectorContent(t *testing.T) {
	const fullValue = "first line\nsecond line with complete diagnostic detail"
	row := Row{ID: 52, Cols: []string{fullValue}, OriginalIndex: 8}
	column := ui.ColumnMeta{Name: "message", Index: 0, Visible: true, Width: 16}

	tableRow, height := ui.RenderRow(ui.RowRenderInput{
		Cols:          row.Cols,
		OriginalIndex: row.OriginalIndex,
		TotalRows:     1,
		ColsMeta:      []ui.ColumnMeta{column},
		ContentWidth:  16,
		Styles: ui.RowStyles{
			Cell:          lipgloss.NewStyle().Padding(0, 1),
			DefaultMarker: " ",
		},
	})
	if height != 1 || strings.Count(tableRow, "\n") != 0 || !strings.Contains(tableRow, "↵") || !strings.Contains(tableRow, "…") {
		t.Fatalf("table row is not a one-line summary: height=%d row=%q", height, tableRow)
	}

	m := Model{table: tableState{header: []ui.ColumnMeta{column}}}
	inspector, _ := m.buildInspectorContent(row, 0, 30)
	plain := stripANSI(inspector)
	normalized := strings.Join(strings.Fields(plain), " ")
	if !strings.Contains(normalized, "first line") || !strings.Contains(normalized, "second line with complete diagnostic detail") {
		t.Fatalf("inspector lost full multiline content:\n%s", plain)
	}
}

func TestRowInspectorFieldNavigationWrapsAndMovesViewport(t *testing.T) {
	m := inspectorTestModel()
	m.inspectorPort = viewport.New(32, 1)
	m.refreshInspectorContent()
	if got := m.inspectorPort.YOffset; got != 0 {
		t.Fatalf("first field should retain metadata at top, got offset %d", got)
	}

	m.cycleInspectorField(-1)
	if got := m.view.inspector.selectedField; got != 2 {
		t.Fatalf("previous field should wrap to last field, got %d", got)
	}
	m.refreshInspectorContent()
	if got := m.inspectorPort.YOffset; got == 0 {
		t.Fatal("selecting the last field should bring it into view")
	}

	m.cycleInspectorField(1)
	if got := m.view.inspector.selectedField; got != 0 {
		t.Fatalf("next field should wrap to first field, got %d", got)
	}
}

func TestRowInspectorViewMatchesPanelWidthAndLayoutBudget(t *testing.T) {
	const panelWidth = 96
	m := inspectorTestModel()
	m.view.inspector.open = true
	m.inspectorPort = viewport.New(0, 0)
	m.terminalHeight = 30
	m.terminalWidth = panelWidth
	m.recomputeLayout(m.terminalHeight, m.terminalWidth)
	m.refreshInspectorContent()

	if got, want := m.inspectorPort.Height, 2; got != want {
		t.Fatalf("unexpected inspector content height: got %d want %d", got, want)
	}
	if got, want := m.viewport.Height, 20; got != want {
		t.Fatalf("unexpected table viewport height: got %d want %d", got, want)
	}

	lines := strings.Split(stripANSI(m.rowInspectorView(panelWidth)), "\n")
	if got, want := len(lines), 2+inspectorChromeRows; got != want {
		t.Fatalf("inspector line count mismatch: got %d want %d", got, want)
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got != panelWidth {
			t.Fatalf("line %d width mismatch: got %d want %d (%q)", i, got, panelWidth, line)
		}
	}
	if !strings.Contains(lines[0], "Source 7") {
		t.Fatalf("inspector border missing original row: %q", lines[0])
	}
}

func TestRowInspectorPacksShortFieldsAndKeepsLongFieldsFullWidth(t *testing.T) {
	m := inspectorTestModel()
	m.table.header = append(m.table.header, ui.ColumnMeta{Name: "message", Index: 3, Visible: true})
	m.table.rows[0].Cols = append(m.table.rows[0].Cols, strings.Repeat("long detail ", 12))

	content, offsets := m.buildInspectorContent(m.table.rows[0], 0, 96)
	plain := stripANSI(content)
	lines := strings.Split(plain, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected packed fields plus wrapped detail, got:\n%s", plain)
	}
	if !strings.Contains(lines[0], "first") || !strings.Contains(lines[0], "second") {
		t.Fatalf("short fields were not packed into two lanes: %q", lines[0])
	}
	if offsets[0] != offsets[1] {
		t.Fatalf("packed fields should share a row offset: %v", offsets)
	}
	if offsets[3] <= offsets[2] {
		t.Fatalf("long field should start on its own row: %v", offsets)
	}
}

func TestRowInspectorLabelsUnnamedColumns(t *testing.T) {
	m := inspectorTestModel()
	m.table.header[0].Name = ""
	content, _ := m.buildInspectorContent(m.table.rows[0], 0, 96)
	if !strings.Contains(stripANSI(content), "Column 1  one") {
		t.Fatalf("unnamed column did not receive a stable display label: %q", stripANSI(content))
	}
	name, _, ok := m.inspectorFieldClipboardText()
	if !ok || name != "Column 1" {
		t.Fatalf("unexpected clipboard field label: name=%q ok=%t", name, ok)
	}
}

func TestRowInspectorUsesRemainingSpaceOnShortTerminal(t *testing.T) {
	m := inspectorTestModel()
	m.view.inspector.open = true
	m.recomputeLayout(10, 80)

	if got := m.inspectorPort.Height; got != 1 {
		t.Fatalf("short layout should reduce inspector to one content row, got %d", got)
	}
	if got := m.viewport.Height; got != 1 {
		t.Fatalf("short layout should retain one table row, got %d", got)
	}
}

func TestRowInspectorHeightTracksCurrentContent(t *testing.T) {
	m := inspectorTestModel()
	m.view.inspector.open = true
	m.terminalHeight = 30
	m.terminalWidth = 96
	m.recomputeLayout(m.terminalHeight, m.terminalWidth)
	if got, want := m.inspectorPort.Height, 2; got != want {
		t.Fatalf("unexpected compact height: got %d want %d", got, want)
	}

	m.table.commentRows[m.table.rows[0].ID] = strings.Repeat("long comment content ", 20)
	m.refreshView("comment-added", false)
	if got, want := m.inspectorPort.Height, inspectorMaxContentRows; got != want {
		t.Fatalf("inspector did not grow for long content: got %d want %d", got, want)
	}

	delete(m.table.commentRows, m.table.rows[0].ID)
	m.refreshView("comment-removed", false)
	if got, want := m.inspectorPort.Height, 2; got != want {
		t.Fatalf("inspector did not shrink after content removal: got %d want %d", got, want)
	}
}

func inspectorTestModel() Model {
	row := Row{ID: 91, Cols: []string{"one", "two", "three"}, OriginalIndex: 7}
	return Model{
		cursor: 0,
		table: tableState{
			header: []ui.ColumnMeta{
				{Name: "first", Index: 0, Visible: true},
				{Name: "second", Index: 1, Visible: true},
				{Name: "third", Index: 2, Visible: true},
			},
			rows:            []Row{row},
			filteredIndices: []int{0},
			markedRows:      map[uint64]ui.MarkColor{},
			commentRows:     map[uint64]string{},
		},
		view: viewState{inspector: rowInspectorState{lastRowIndex: -1, lastField: -1}},
	}
}
