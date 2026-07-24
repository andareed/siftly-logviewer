package siftly

import (
	"testing"
	"time"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func TestUndoRedoReturnsDirtyStateToSavedBaseline(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.markCurrent(ui.MarkGreen)
	if !m.dirty || !m.canUndo() {
		t.Fatalf("mark should create an undoable dirty change: dirty=%t undo=%t", m.dirty, m.canUndo())
	}

	_ = m.undoLastChange()
	if m.dirty {
		t.Fatal("undoing to the initial baseline should clear dirty state")
	}
	if _, exists := m.table.markedRows[m.table.rows[0].ID]; exists {
		t.Fatal("undo did not remove the mark")
	}
	if !m.canRedo() {
		t.Fatal("undo should create a redo entry")
	}

	_ = m.redoLastChange()
	if !m.dirty || m.table.markedRows[m.table.rows[0].ID] != ui.MarkGreen {
		t.Fatalf("redo did not restore mark and dirty state: dirty=%t mark=%q", m.dirty, m.table.markedRows[m.table.rows[0].ID])
	}

	m.markSavedBaseline()
	if m.dirty {
		t.Fatal("successful save baseline should be clean")
	}
	m.addComment("watch this")
	if !m.dirty {
		t.Fatal("comment after save should be dirty")
	}
	_ = m.undoLastChange()
	if m.dirty || m.table.commentRows[m.table.rows[0].ID] != "" {
		t.Fatalf("undo should return to saved baseline: dirty=%t comment=%q", m.dirty, m.table.commentRows[m.table.rows[0].ID])
	}
}

func TestViewOperationsAreUndoableWithoutMakingDocumentDirty(t *testing.T) {
	m := newChangeTrackingTestModel()
	if err := m.setSortSpec("second desc"); err != nil {
		t.Fatalf("set sort: %v", err)
	}
	if m.dirty {
		t.Fatal("sort is a view operation and should not make persisted data dirty")
	}
	if !m.table.sortEnabled || !m.canUndo() {
		t.Fatal("sort should be active and undoable")
	}

	_ = m.undoLastChange()
	if m.table.sortEnabled {
		t.Fatal("undo did not restore unsorted view")
	}
	_ = m.redoLastChange()
	if !m.table.sortEnabled || !m.table.sortDesc {
		t.Fatal("redo did not restore sorted view")
	}

	if err := m.setFilterPattern("alpha"); err != nil {
		t.Fatalf("set filter: %v", err)
	}
	if m.dirty {
		t.Fatal("filter is a view operation and should not make persisted data dirty")
	}
	_ = m.undoLastChange()
	if m.table.filterPattern != "" || m.table.filterEnabled {
		t.Fatalf("undo did not clear filter: pattern=%q enabled=%t", m.table.filterPattern, m.table.filterEnabled)
	}
}

func TestFrozenColumnIsTrackedByUndoAndDirtyState(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.table.header[0].WrapLines = 4
	m.initializeChangeTracking()
	m.table.header[0].Frozen = true
	m.recordChange("freeze column")
	m.view.columnScrollOffset = 4
	if !m.dirty || !m.canUndo() {
		t.Fatalf("freeze should be dirty and undoable: dirty=%t undo=%t", m.dirty, m.canUndo())
	}

	_ = m.undoLastChange()
	if m.table.header[0].Frozen || m.table.header[0].WrapLines != 4 || m.dirty || m.view.columnScrollOffset != 0 {
		t.Fatalf("undo freeze: column=%+v dirty=%t offset=%d", m.table.header[0], m.dirty, m.view.columnScrollOffset)
	}
	_ = m.redoLastChange()
	if !m.table.header[0].Frozen || m.table.header[0].WrapLines != 4 || !m.dirty {
		t.Fatalf("redo freeze: column=%+v dirty=%t", m.table.header[0], m.dirty)
	}
}

func TestLegacyRecoveryStateRetainsCurrentSchemaWrapping(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.table.header[0].WrapLines = 4
	state := m.captureTrackedState()
	state.Columns[0].WrapLines = 0

	m.restoreTrackedState(state, false)
	if got := m.table.header[0].WrapLines; got != 4 {
		t.Fatalf("legacy recovery removed schema wrapping: WrapLines=%d", got)
	}
}

func TestQuitRequiresSecondPressWhenDirty(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.markCurrent(ui.MarkAmber)

	first := m.confirmOrQuit()
	if first == nil || !m.quitConfirmationActive() {
		t.Fatal("first quit should show confirmation instead of quitting")
	}
	if m.view.notice.Msg == "" {
		t.Fatal("first quit should display an unsaved warning")
	}

	second := m.confirmOrQuit()
	if second == nil {
		t.Fatal("confirmed quit should return tea.Quit")
	}
	quitMsg := second()
	if _, ok := quitMsg.(tea.QuitMsg); !ok {
		t.Fatalf("confirmed quit returned %T, want tea.QuitMsg", quitMsg)
	}

	m.view.quitConfirmUntil = time.Now().Add(time.Second)
	if action := m.resolveViewAction(tea.KeyMsg{Type: tea.KeyEsc}); action != viewActionCancelQuit {
		t.Fatalf("escape action = %v want cancel quit", action)
	}
}

func TestNormalModeHistoryAndPagingKeys(t *testing.T) {
	m := newChangeTrackingTestModel()

	tests := []struct {
		name string
		msg  tea.KeyMsg
		want viewAction
	}{
		{name: "undo", msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}, want: viewActionUndo},
		{name: "redo", msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}, want: viewActionRedo},
		{name: "page up", msg: tea.KeyMsg{Type: tea.KeyCtrlU}, want: viewActionPageUp},
		{name: "page down", msg: tea.KeyMsg{Type: tea.KeyCtrlD}, want: viewActionPageDown},
	}
	for _, tt := range tests {
		if got := m.resolveViewAction(tt.msg); got != tt.want {
			t.Fatalf("%s action = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func newChangeTrackingTestModel() *Model {
	rows := []Row{
		{ID: 11, Cols: []string{"alpha", "2"}, OriginalIndex: 1},
		{ID: 12, Cols: []string{"beta", "1"}, OriginalIndex: 2},
	}
	m := &Model{
		cursor: 0,
		table: tableState{
			header: []ui.ColumnMeta{
				{Name: "first", Index: 0, Visible: true, MinWidth: 8, Weight: 1},
				{Name: "second", Index: 1, Visible: true, MinWidth: 8, Weight: 1},
			},
			rows:            rows,
			filteredIndices: []int{0, 1},
			markedRows:      map[uint64]ui.MarkColor{},
			commentRows:     map[uint64]string{},
			sortColumn:      -1,
			rowOrder:        []int{0, 1},
			searchColumns:   []int{0, 1},
		},
	}
	m.InitialiseView()
	m.applyFilter()
	return m
}
