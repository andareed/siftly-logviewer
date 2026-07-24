package ui

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestDefaultStylesUseSharedDesignTokens(t *testing.T) {
	styles := DefaultStyles()
	tokens := styles.ResolvedTokens()

	assertTerminalColor(t, "header foreground", styles.Header.GetForeground(), tokens.Colors.TextStrong)
	assertTerminalColor(t, "header background", styles.Header.GetBackground(), tokens.Colors.SurfaceHeader)
	assertTerminalColor(t, "selected foreground", styles.RowSelected.GetForeground(), tokens.Colors.TextSelected)
	assertTerminalColor(t, "selected background", styles.RowSelected.GetBackground(), tokens.Colors.SurfaceSelected)
	assertTerminalColor(t, "range background", styles.RowRangeSelectedBG, tokens.Colors.SurfaceRange)
	assertTerminalColor(t, "repeated text", styles.RepeatedCell.GetForeground(), tokens.Colors.TextSubtle)
	assertTerminalColor(t, "search foreground", styles.SearchHighlight.GetForeground(), tokens.Colors.SearchForeground)
	assertTerminalColor(t, "search background", styles.SearchHighlight.GetBackground(), tokens.Colors.SearchBackground)
	assertTerminalColor(t, "table border", styles.Table.GetBorderTopForeground(), tokens.Colors.Border)
}

func TestFooterAndNoticesUseSharedDesignTokens(t *testing.T) {
	tokens := DefaultDesignTokens()
	footer := FooterStylesFromTokens(tokens)

	assertTerminalColor(t, "footer background", footer.BarBG, tokens.Colors.SurfaceFooter)
	assertTerminalColor(t, "footer inset", footer.StatusBG, tokens.Colors.SurfaceFooterInset)
	assertTerminalColor(t, "footer success", footer.SuccessFG, tokens.Colors.Success)
	assertTerminalColor(t, "footer warning", footer.WarnFG, tokens.Colors.Warning)
	assertTerminalColor(t, "footer error", footer.ErrorFG, tokens.Colors.Error)
	assertTerminalColor(t, "notice success", tokens.NoticeStyle("success").GetForeground(), tokens.Colors.Success)
	assertTerminalColor(t, "notice warning", tokens.NoticeStyle("warn").GetForeground(), tokens.Colors.Warning)
	assertTerminalColor(t, "notice error", tokens.NoticeStyle("error").GetForeground(), tokens.Colors.Error)
}

func TestDefaultInteractionStatesRemainDistinct(t *testing.T) {
	tokens := DefaultDesignTokens()
	if reflect.DeepEqual(tokens.States.Selected.GetBackground(), tokens.States.Range.GetBackground()) {
		t.Fatal("selected and range states must use distinct backgrounds")
	}
	if reflect.DeepEqual(tokens.Emphasis.Normal.GetForeground(), tokens.Emphasis.Muted.GetForeground()) {
		t.Fatal("normal and muted emphasis must use distinct foregrounds")
	}
	if reflect.DeepEqual(tokens.Emphasis.Muted.GetForeground(), tokens.Emphasis.Subtle.GetForeground()) {
		t.Fatal("muted and subtle emphasis must use distinct foregrounds")
	}
}

func TestDefaultDesignTokensReturnsIndependentGraphPalettes(t *testing.T) {
	first := DefaultDesignTokens()
	second := DefaultDesignTokens()
	first.GraphSeries[0] = first.Colors.Text
	if reflect.DeepEqual(first.GraphSeries, second.GraphSeries) {
		t.Fatal("default token calls must not share a mutable graph palette")
	}
}

func TestDefaultDesignTokensUsesExplicitANSI256Palette(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(previousProfile)

	tokens := DefaultDesignTokens()
	assertTerminalColor(t, "selected ANSI-256 background", tokens.Colors.SurfaceSelected, lipgloss.Color("238"))
	assertTerminalColor(t, "range ANSI-256 background", tokens.Colors.SurfaceRange, lipgloss.Color("235"))
	assertTerminalColor(t, "normal ANSI-256 text", tokens.Colors.Text, lipgloss.Color("250"))
	assertTerminalColor(t, "ANSI-256 accent", tokens.Colors.Accent, lipgloss.Color("214"))

	if reflect.DeepEqual(tokens.Colors.SurfaceSelected, tokens.Colors.SurfaceRange) {
		t.Fatal("ANSI-256 cursor and range backgrounds must remain distinct")
	}
}

func TestANSI256PaletteContainsOnlyExplicitIndexes(t *testing.T) {
	colors, graphSeries := ansi256Palette()
	value := reflect.ValueOf(colors)
	valueType := value.Type()

	for i := 0; i < value.NumField(); i++ {
		name := valueType.Field(i).Name
		assertANSI256Index(t, name, string(value.Field(i).Interface().(lipgloss.Color)))
	}
	for i, color := range graphSeries {
		assertANSI256Index(t, "graph series "+strconv.Itoa(i), string(color))
	}
}

func TestDefaultStylesRenderDistinctANSI256RowBackgrounds(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(previousProfile)

	styles := DefaultStyles()
	rowStyles := RowStyles{
		Row:                styles.Row,
		RowSelected:        styles.RowSelected,
		Cell:               styles.Cell,
		RowTextFGColor:     styles.RowTextFGColor,
		RowSelectedFG:      styles.RowSelectedFG,
		RowSelectedBG:      styles.RowSelectedBG,
		RowRangeSelectedBG: styles.RowRangeSelectedBG,
		DefaultMarker:      styles.DefaultMarker,
	}
	input := RowRenderInput{
		Cols:          []string{"row"},
		OriginalIndex: 1,
		TotalRows:     2,
		ColsMeta:      []ColumnMeta{{Visible: true, Width: 8}},
		Styles:        rowStyles,
	}

	input.Selected = true
	cursorRow, _ := RenderRow(input)
	if !strings.Contains(cursorRow, "48;5;238") || strings.Contains(cursorRow, "48;5;235") {
		t.Fatalf("cursor row does not use only ANSI-256 background 238: %q", cursorRow)
	}

	input.Selected = false
	input.RangeSelected = true
	rangeRow, _ := RenderRow(input)
	if !strings.Contains(rangeRow, "48;5;235") || strings.Contains(rangeRow, "48;5;238") {
		t.Fatalf("range row does not use only ANSI-256 background 235: %q", rangeRow)
	}
}

func assertANSI256Index(t *testing.T, name, value string) {
	t.Helper()
	index, err := strconv.Atoi(value)
	if err != nil || index < 0 || index > 255 {
		t.Fatalf("%s ANSI-256 colour = %q, want an explicit index from 0 to 255", name, value)
	}
}
