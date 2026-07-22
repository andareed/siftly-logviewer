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
