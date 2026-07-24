package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderRowCellsUsesColumnSourceIndices(t *testing.T) {
	got, _ := RenderRowCells(
		[]string{"left", "right"},
		[]ColumnMeta{
			{Name: "second", Index: 1, Visible: true, Width: 5},
			{Name: "first", Index: 0, Visible: true, Width: 5},
		},
		lipgloss.NewStyle(),
	)
	if !strings.HasPrefix(got, "rightleft") {
		t.Fatalf("reordered cells = %q, want right then left", got)
	}
}

func TestScrollableRowCellsPinsFrozenSegment(t *testing.T) {
	columns := []ColumnMeta{
		{Name: "fixed", Index: 0, Visible: true, Frozen: true, Width: 4},
		{Name: "moving", Index: 1, Visible: true, Width: 4},
	}
	got, _ := renderScrollableRowCells(
		[]string{"AAAA", "1234"},
		columns,
		lipgloss.NewStyle(),
		rowCellOptions{},
		7,
		2,
	)
	if got != "AAAA│34" {
		t.Fatalf("scrolled cells = %q, want %q", got, "AAAA│34")
	}
}

func TestScrollableWrappedRowPadsFrozenSegmentAcrossContinuationLines(t *testing.T) {
	columns := []ColumnMeta{
		{Name: "fixed", Index: 0, Visible: true, Frozen: true, Width: 4},
		{Name: "moving", Index: 1, Visible: true, Width: 6, WrapLines: 3},
	}
	got, height := renderScrollableRowCells(
		[]string{"AAAA", "one two three four"},
		columns,
		lipgloss.NewStyle(),
		rowCellOptions{},
		9,
		0,
	)
	if height != 3 {
		t.Fatalf("wrapped scrolled row height=%d want 3: %q", height, got)
	}
	lines := strings.Split(got, "\n")
	if lines[0] != "AAAA│one " {
		t.Fatalf("first wrapped line=%q", lines[0])
	}
	for i, line := range lines[1:] {
		if !strings.HasPrefix(line, "    │") {
			t.Fatalf("continuation line %d did not retain frozen segment: %q", i+1, line)
		}
	}
}
