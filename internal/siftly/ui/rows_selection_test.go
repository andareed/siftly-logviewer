package ui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestResolveRowVisualStyleKeepsCursorDistinctFromRange(t *testing.T) {
	styles := RowStyles{
		RowTextFGColor:     lipgloss.Color("normal-fg"),
		RowSelectedFG:      lipgloss.Color("cursor-fg"),
		RowSelectedBG:      lipgloss.Color("cursor-bg"),
		RowRangeSelectedBG: lipgloss.Color("range-bg"),
	}

	_, fg, bg := resolveRowVisualStyle(styles, false, true)
	assertTerminalColor(t, "range foreground", fg, styles.RowTextFGColor)
	assertTerminalColor(t, "range background", bg, styles.RowRangeSelectedBG)

	_, fg, bg = resolveRowVisualStyle(styles, true, true)
	assertTerminalColor(t, "cursor foreground", fg, styles.RowSelectedFG)
	assertTerminalColor(t, "cursor background", bg, styles.RowSelectedBG)
}

func TestRenderRowEmitsRangeBackground(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	styles := RowStyles{
		Row:                lipgloss.NewStyle(),
		RowSelected:        lipgloss.NewStyle().Background(lipgloss.Color("#3a3a3a")),
		Cell:               lipgloss.NewStyle(),
		RowTextFGColor:     lipgloss.Color("#c0c0c0"),
		RowSelectedFG:      lipgloss.Color("#e0e0e0"),
		RowSelectedBG:      lipgloss.Color("#3a3a3a"),
		RowRangeSelectedBG: lipgloss.Color("#2a2a2a"),
		DefaultMarker:      " ",
	}

	got, _ := RenderRow(RowRenderInput{
		Cols:          []string{"range row"},
		OriginalIndex: 1,
		RangeSelected: true,
		TotalRows:     2,
		ColsMeta: []ColumnMeta{
			{Visible: true, Width: 12},
		},
		Styles: styles,
	})
	if count := strings.Count(got, "48;2;42;42;42"); count < 2 {
		t.Fatalf("rendered range row applies #2a2a2a background %d times, want marker and row body: %q", count, got)
	}
}

func assertTerminalColor(t *testing.T, name string, got, want lipgloss.TerminalColor) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %#v want %#v", name, got, want)
	}
}
