package siftly

import (
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestFullSourceReloadPreservesAnnotationsForMatchingRows(t *testing.T) {
	t.Parallel()

	kept := Row{Cols: []string{"keep", "metric"}}
	kept.ID = kept.ComputeID()
	kept.OriginalIndex = 1
	fullOnly := Row{Cols: []string{"full", "metric"}}
	fullOnly.ID = fullOnly.ComputeID()
	fullOnly.OriginalIndex = 2

	m := &Model{
		ready:          true,
		terminalHeight: 24,
		terminalWidth:  80,
		table: tableState{
			rows:        []Row{kept},
			markedRows:  map[uint64]ui.MarkColor{kept.ID: ui.MarkAmber},
			commentRows: map[uint64]string{kept.ID: "watch this"},
		},
	}
	m.view.operation.active = true
	m.view.operation.seq = 42

	full := &Model{
		ready:          true,
		terminalHeight: 24,
		terminalWidth:  80,
		table: tableState{
			rows:        []Row{fullOnly, kept},
			markedRows:  map[uint64]ui.MarkColor{},
			commentRows: map[uint64]string{},
		},
	}

	_ = m.handleFullSourceReloadComplete(fullSourceReloadCompleteMsg{
		ID:             42,
		Model:          full,
		CurrentRowHash: kept.ID,
	})

	if got, want := len(m.table.rows), 2; got != want {
		t.Fatalf("row count = %d want %d", got, want)
	}
	if got := m.table.markedRows[kept.ID]; got != ui.MarkAmber {
		t.Fatalf("mark = %q want %q", got, ui.MarkAmber)
	}
	if got := m.table.commentRows[kept.ID]; got != "watch this" {
		t.Fatalf("comment = %q want watch this", got)
	}
	if m.CanReloadFullSource() {
		t.Fatalf("full dataset should not keep prefilter reload hook")
	}
}
