package siftly

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestApplyColumnManagerResultUpdatesLayoutAndDataSort(t *testing.T) {
	m, err := NewModelFromRecords([][]string{
		{"time", "message", "value"},
		{"one", "short", "9"},
		{"two", "a much longer message", "20"},
	}, ColumnSchema{})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.InitialiseView()
	m.ready = true
	m.terminalWidth = 80
	m.terminalHeight = 24
	m.recomputeLayout(m.terminalHeight, m.terminalWidth)

	err = m.applyColumnManagerResult(dialogs.ColumnManagerResult{
		Columns: []dialogs.ColumnManagerItem{
			{SourceIndex: 2, Name: "value", Visible: true, Frozen: true},
			{SourceIndex: 0, Name: "time", Visible: true},
			{SourceIndex: 1, Name: "message", Visible: true, AutoFit: true},
		},
		SortEnabled: true,
		SortColumn:  2,
		SortDesc:    true,
	})
	if err != nil {
		t.Fatalf("apply manager result: %v", err)
	}

	if got := m.table.header[0]; got.Index != 2 || !got.Frozen {
		t.Fatalf("first column = %+v, want frozen value column", got)
	}
	message := m.table.header[2]
	if message.Index != 1 || message.MinWidth != len("a much longer message") || message.Weight != 0 {
		t.Fatalf("auto-fit message column = %+v", message)
	}
	if got := m.table.rowOrder[0]; got != 1 {
		t.Fatalf("descending numeric row order starts at %d, want row 1", got)
	}
	if !m.dirty {
		t.Fatal("column layout should mark model dirty")
	}
}

func TestAutoFitUsesDisplayedRows(t *testing.T) {
	m := &Model{
		table: tableState{
			header: []ui.ColumnMeta{{Name: "message", Index: 0, Visible: true}},
			rows: []Row{
				{Cols: []string{"short"}},
				{Cols: []string{"a value outside the current filter"}},
			},
			filteredIndices: []int{0},
		},
	}
	if got := m.autoFitColumnWidth(m.table.header[0], false); got != len("message") {
		t.Fatalf("auto-fit width = %d, want displayed/header width %d", got, len("message"))
	}
}

func TestColumnScrollOffsetAccountsForFrozenWidth(t *testing.T) {
	m := &Model{
		viewport: viewportForColumnTest(14),
		table: tableState{
			header: []ui.ColumnMeta{
				{Index: 0, Visible: true, Frozen: true, Width: 4},
				{Index: 1, Visible: true, Width: 12},
			},
			rows: []Row{{Cols: []string{"fixed", "scrolling"}}},
		},
		styles: ui.Styles{
			Cell:          lipgloss.NewStyle(),
			PillMarker:    "|",
			CommentMarker: "",
		},
	}
	if maximum := m.maxColumnScrollOffset(); maximum <= 0 {
		t.Fatalf("maximum scroll offset = %d, want positive", maximum)
	}
	m.scrollColumns(1 << 20)
	if m.view.columnScrollOffset != m.maxColumnScrollOffset() {
		t.Fatalf("clamped scroll offset = %d want %d", m.view.columnScrollOffset, m.maxColumnScrollOffset())
	}
}

func viewportForColumnTest(width int) viewport.Model {
	result := viewport.New(width, 1)
	return result
}

func TestColumnManagerAliasesOpenSameDialog(t *testing.T) {
	for _, alias := range []rune{'c', 's', 'o'} {
		m := newChangeTrackingTestModel()
		m.terminalWidth = 100
		m.terminalHeight = 30
		_, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
		_, _ = m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{alias}})
		if _, ok := m.activeDialog.(*dialogs.ColumnManager); !ok {
			t.Fatalf("v %c opened %T, want column manager", alias, m.activeDialog)
		}
	}
}

func TestExportModelUsesDisplayOrderValues(t *testing.T) {
	m, err := NewModelFromRecords([][]string{
		{"first", "second"},
		{"A", "B"},
	}, ColumnSchema{})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.table.header = []ui.ColumnMeta{m.table.header[1], m.table.header[0]}
	m.table.filteredIndices = []int{0}

	path := filepath.Join(t.TempDir(), "ordered.csv")
	if err := ExportModel(m, path); err != nil {
		t.Fatalf("export: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open export: %v", err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if records[0][0] != "second" || records[0][1] != "first" {
		t.Fatalf("header order = %v", records[0])
	}
	if records[1][0] != "B" || records[1][1] != "A" {
		t.Fatalf("value order = %v", records[1])
	}
}
