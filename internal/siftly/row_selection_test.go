package siftly

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestRangeSelectionKeyFlow(t *testing.T) {
	m := newRowSelectionTestModel()
	m.cursor = 1

	_, _ = m.handleViewModeKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.view.rowRange.active {
		t.Fatal("space should start a row selection")
	}

	_, _ = m.handleViewModeKey(tea.KeyMsg{Type: tea.KeyDown})
	if got := m.selectedRowCount(); got != 2 {
		t.Fatalf("selected row count = %d want 2", got)
	}
	_, _ = m.handleViewModeKey(tea.KeyMsg{Type: tea.KeyDown})
	_, _ = m.handleViewModeKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != len(m.table.filteredIndices)-1 {
		t.Fatalf("cursor = %d want filtered-row end %d", m.cursor, len(m.table.filteredIndices)-1)
	}

	_, _ = m.handleViewModeKey(tea.KeyMsg{Type: tea.KeyEsc})
	if m.view.rowRange.active {
		t.Fatal("escape should clear the row selection")
	}
}

func TestApplyingFilterClearsRangeSelection(t *testing.T) {
	m := newRowSelectionTestModel()
	m.toggleRowRangeSelection()
	m.applyFilter()
	if m.view.rowRange.active {
		t.Fatal("applying a filter should clear the ephemeral row selection")
	}
}

func TestSelectedRowsClipboardTextUsesFilteredDisplayOrder(t *testing.T) {
	m := newRowSelectionTestModel()
	m.cursor = 2
	m.toggleRowRangeSelection()
	m.cursor = 0

	text, count, err := m.selectedRowsClipboardText()
	if err != nil {
		t.Fatalf("clipboard text: %v", err)
	}
	if count != 3 {
		t.Fatalf("clipboard row count = %d want 3", count)
	}
	want := "row-3\tvalue-3\nrow-1\tvalue-1\nrow-4\tvalue-4"
	if text != want {
		t.Fatalf("clipboard text = %q want %q", text, want)
	}
}

func TestSelectedRowsClipboardTextRejectsVeryLargeRange(t *testing.T) {
	m := &Model{
		cursor: maxClipboardRows,
		table: tableState{
			filteredIndices: make([]int, maxClipboardRows+1),
		},
		view: viewState{
			rowRange: rowRangeSelection{active: true, anchor: 0},
		},
	}

	_, _, err := m.selectedRowsClipboardText()
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("large selection error = %v", err)
	}
}

func TestSelectedRowsClipboardTextSeparatesEmptyRows(t *testing.T) {
	m := &Model{
		cursor: 1,
		table: tableState{
			rows:            []Row{{Cols: []string{""}}, {Cols: []string{"next"}}},
			filteredIndices: []int{0, 1},
		},
		view: viewState{
			rowRange: rowRangeSelection{active: true, anchor: 0},
		},
	}

	text, count, err := m.selectedRowsClipboardText()
	if err != nil {
		t.Fatalf("clipboard text: %v", err)
	}
	if count != 2 || text != "\nnext" {
		t.Fatalf("clipboard result = (%q, %d) want (%q, 2)", text, count, "\nnext")
	}
}

func TestMarkCommandAppliesToActiveRange(t *testing.T) {
	m := newRowSelectionTestModel()
	m.cursor = 0
	m.toggleRowRangeSelection()
	m.cursor = 2
	_ = m.enterCommand(CmdMark, "", false, false)

	_, _ = m.handleMarkCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

	for _, rowIndex := range m.table.filteredIndices {
		if got := m.table.markedRows[m.table.rows[rowIndex].ID]; got != ui.MarkGreen {
			t.Fatalf("displayed row %d mark = %q want %q", rowIndex, got, ui.MarkGreen)
		}
	}
	if got := m.table.markedRows[m.table.rows[1].ID]; got != ui.MarkNone {
		t.Fatalf("filtered-out row mark = %q want none", got)
	}
}

func TestPanelStatusShowsSelectedRowCount(t *testing.T) {
	top := renderPanelTopBorder(
		"hostlog.json",
		panelStatusSpec{CurrentRow: 4, TotalRows: 20, Selected: 3},
		60,
	)
	if !strings.Contains(top, "Selected: 3") {
		t.Fatalf("selection count missing from panel status: %q", top)
	}
}

func newRowSelectionTestModel() *Model {
	rows := make([]Row, 4)
	for i := range rows {
		rows[i] = Row{
			Cols:          []string{"row-" + string(rune('1'+i)), "value-" + string(rune('1'+i))},
			OriginalIndex: i + 1,
		}
		rows[i].ID = rows[i].ComputeID()
	}
	return &Model{
		cursor: 0,
		table: tableState{
			rows:            rows,
			filteredIndices: []int{2, 0, 3},
			markedRows:      make(map[uint64]ui.MarkColor),
		},
		view: viewState{mode: modeView},
	}
}
