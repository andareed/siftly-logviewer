package ui

import (
	"reflect"
	"testing"
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
