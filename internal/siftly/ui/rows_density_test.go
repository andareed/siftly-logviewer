package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestRenderRowCellsCompactsMultilineValuesToOneLine(t *testing.T) {
	got, height := RenderRowCells(
		[]string{"alpha\nbeta gamma"},
		[]ColumnMeta{{Index: 0, Visible: true, Width: 12}},
		lipgloss.NewStyle().Padding(0, 1),
	)
	if height != 1 || strings.Contains(got, "\n") {
		t.Fatalf("dense row height=%d output=%q", height, got)
	}
	if !strings.Contains(got, "↵") || !strings.Contains(got, "…") {
		t.Fatalf("dense row should signal line break and truncation: %q", got)
	}
	if width := lipgloss.Width(got); width != 12 {
		t.Fatalf("dense cell width=%d want 12 (%q)", width, got)
	}
}

func TestRenderRowCellsWrapsEnabledColumnWithinLineBudget(t *testing.T) {
	got, height := RenderRowCells(
		[]string{"alpha\nbeta gamma delta epsilon"},
		[]ColumnMeta{{Index: 0, Visible: true, Width: 12, WrapLines: 2}},
		lipgloss.NewStyle().Padding(0, 1),
	)
	if height != 2 {
		t.Fatalf("wrapped row height=%d want 2: %q", height, got)
	}
	if !strings.Contains(got, "\n") || strings.Contains(got, "↵") {
		t.Fatalf("wrapped row should preserve visual lines without a line-break marker: %q", got)
	}
	if !strings.Contains(got, "…") {
		t.Fatalf("wrapped row should signal content beyond its line budget: %q", got)
	}
	for i, line := range strings.Split(got, "\n") {
		if width := lipgloss.Width(line); width != 12 {
			t.Fatalf("wrapped line %d width=%d want 12 (%q)", i, width, line)
		}
	}
}

func TestRepeatedCellEmphasisYieldsToSelectionAndSearch(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	styles := RowStyles{
		Row:             lipgloss.NewStyle(),
		RowSelected:     lipgloss.NewStyle(),
		RepeatedCell:    lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")),
		Cell:            lipgloss.NewStyle().Padding(0, 1),
		SearchHighlight: lipgloss.NewStyle().Background(lipgloss.Color("#f5c542")),
		RowTextFGColor:  lipgloss.Color("#c0c0c0"),
		RowSelectedFG:   lipgloss.Color("#e0e0e0"),
		RowSelectedBG:   lipgloss.Color("#3a3a3a"),
		DefaultMarker:   " ",
	}
	input := RowRenderInput{
		Cols:          []string{"same"},
		OriginalIndex: 2,
		TotalRows:     2,
		ColsMeta:      []ColumnMeta{{Index: 0, Visible: true, Width: 8}},
		RepeatedCols:  []bool{true},
		ContentWidth:  8,
		Styles:        styles,
	}

	repeated, _ := RenderRow(input)
	if !strings.Contains(repeated, "38;2;119;119;119") {
		t.Fatalf("repeated row missing quiet foreground: %q", repeated)
	}

	input.Selected = true
	selected, _ := RenderRow(input)
	if strings.Contains(selected, "38;2;119;119;119") {
		t.Fatalf("selected row retained repeated-value emphasis: %q", selected)
	}

	input.Selected = false
	input.SearchQuery = "sam"
	searched, _ := RenderRow(input)
	if strings.Contains(searched, "38;2;119;119;119") {
		t.Fatalf("searched row retained repeated-value emphasis: %q", searched)
	}
	if !strings.Contains(searched, "48;2;245;197;") {
		t.Fatalf("searched row missing match highlight: %q", searched)
	}
}
