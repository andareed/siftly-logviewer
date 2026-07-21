package siftly

import (
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestFooterActiveStatesReflectCurrentModel(t *testing.T) {
	m := Model{
		dirty:  true,
		cursor: 1,
		table: tableState{
			header: []ui.ColumnMeta{{Name: "timestamp", Index: 0}},
			rows: []Row{
				{ID: 1, Cols: []string{"one"}},
				{ID: 2, Cols: []string{"two"}},
			},
			filteredIndices: []int{0, 1},
			filterEnabled:   true,
			filterRegex:     mustCompileFilter(t, "error"),
			sortEnabled:     true,
			sortColumn:      0,
			sortDesc:        true,
			showOnlyMarked:  true,
		},
		view: viewState{rowRange: rowRangeSelection{active: true, anchor: 0}},
	}
	m.table.timeWindow.Enabled = true

	got := strings.Join(m.footerActiveStates(), " | ")
	for _, want := range []string{"FILTER", "SORT timestamp ↓", "TIME WINDOW", "MARKS ONLY", "2 SELECTED", "UNSAVED"} {
		if !strings.Contains(got, want) {
			t.Fatalf("footer states missing %q: %q", want, got)
		}
	}
}

func TestMarkAndCommentSetDirtyOnlyWhenDataChanges(t *testing.T) {
	m := Model{
		cursor: 0,
		table: tableState{
			rows:            []Row{{ID: 10, Cols: []string{"row"}}},
			filteredIndices: []int{0},
			markedRows:      map[uint64]ui.MarkColor{},
			commentRows:     map[uint64]string{},
		},
	}

	m.markCurrent(ui.MarkGreen)
	if !m.dirty {
		t.Fatal("marking a row should set dirty state")
	}
	m.dirty = false
	m.markCurrent(ui.MarkGreen)
	if m.dirty {
		t.Fatal("reapplying the same mark should not set dirty state")
	}
	m.addComment("note")
	if !m.dirty {
		t.Fatal("changing a comment should set dirty state")
	}
}

func TestCompletedSearchIndexReportsPosition(t *testing.T) {
	m := Model{
		cursor: 0,
		table: tableState{
			rows: []Row{
				{Cols: []string{"foo"}},
				{Cols: []string{"bar"}},
				{Cols: []string{"another foo"}},
			},
			filteredIndices: []int{0, 1, 2},
		},
	}

	cmd := m.setSearchQuery("foo")
	for cmd != nil {
		msg, ok := cmd().(searchIndexChunkMsg)
		if !ok {
			t.Fatalf("unexpected search index message")
		}
		cmd = m.handleSearchIndexChunk(msg)
	}
	m.cursor = 2
	if got := m.searchStatusLabel(); got != "Match 2/2" {
		t.Fatalf("unexpected search status: %q", got)
	}
}
