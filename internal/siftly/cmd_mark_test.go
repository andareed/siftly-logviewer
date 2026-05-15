package siftly

import (
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func TestMarkCommandAppliesCountToDisplayedRows(t *testing.T) {
	t.Parallel()

	m := newMarkCommandTestModel(5, []int{0, 2, 4}, 0)
	_ = m.enterCommand(CmdMark, "", false, false)

	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})

	for _, rowIdx := range []int{0, 2, 4} {
		if got := m.table.markedRows[m.table.rows[rowIdx].ID]; got != ui.MarkGreen {
			t.Fatalf("row %d mark = %q want %q", rowIdx, got, ui.MarkGreen)
		}
	}
	for _, rowIdx := range []int{1, 3} {
		if got := m.table.markedRows[m.table.rows[rowIdx].ID]; got != ui.MarkNone {
			t.Fatalf("row %d mark = %q want unmarked", rowIdx, got)
		}
	}
	if m.view.mode != modeView {
		t.Fatalf("mode = %v want modeView", m.view.mode)
	}
}

func TestMarkCommandClipsCountAndClearsDisplayedRows(t *testing.T) {
	t.Parallel()

	m := newMarkCommandTestModel(4, []int{0, 1, 2, 3}, 2)
	for _, row := range m.table.rows {
		m.table.markedRows[row.ID] = ui.MarkRed
	}
	_ = m.enterCommand(CmdMark, "", false, false)

	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9")})
	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	if got := m.table.markedRows[m.table.rows[0].ID]; got != ui.MarkRed {
		t.Fatalf("row 0 mark = %q want %q", got, ui.MarkRed)
	}
	if got := m.table.markedRows[m.table.rows[1].ID]; got != ui.MarkRed {
		t.Fatalf("row 1 mark = %q want %q", got, ui.MarkRed)
	}
	for _, rowIdx := range []int{2, 3} {
		if got := m.table.markedRows[m.table.rows[rowIdx].ID]; got != ui.MarkNone {
			t.Fatalf("row %d mark = %q want cleared", rowIdx, got)
		}
	}
}

func TestMarkCommandBackspaceEditsCount(t *testing.T) {
	t.Parallel()

	m := newMarkCommandTestModel(5, []int{0, 1, 2, 3, 4}, 0)
	_ = m.enterCommand(CmdMark, "", false, false)

	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("12")})
	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyBackspace})
	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	if got := m.table.markedRows[m.table.rows[0].ID]; got != ui.MarkAmber {
		t.Fatalf("row 0 mark = %q want %q", got, ui.MarkAmber)
	}
	if got := m.table.markedRows[m.table.rows[1].ID]; got != ui.MarkAmber {
		t.Fatalf("row 1 mark = %q want %q", got, ui.MarkAmber)
	}
	if got := m.table.markedRows[m.table.rows[2].ID]; got != ui.MarkNone {
		t.Fatalf("row 2 mark = %q want unmarked", got)
	}
}

func newMarkCommandTestModel(rowCount int, filteredIndices []int, cursor int) *Model {
	rows := make([]Row, rowCount)
	for i := range rows {
		rows[i] = Row{
			Cols:          []string{string(rune('a' + i))},
			OriginalIndex: i + 1,
		}
		rows[i].ID = rows[i].ComputeID()
	}
	return &Model{
		cursor: cursor,
		table: tableState{
			rows:            rows,
			filteredIndices: append([]int(nil), filteredIndices...),
			markedRows:      make(map[uint64]ui.MarkColor),
		},
	}
}
