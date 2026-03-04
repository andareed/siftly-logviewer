package ui

import (
	"regexp"
	"strings"
	"testing"
)

var ansiFooterPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripFooterANSI(s string) string {
	return ansiFooterPattern.ReplaceAllString(s, "")
}

func TestRenderFooterInputModeShowsPromptMarkerWhenEmpty(t *testing.T) {
	out := stripFooterANSI(RenderFooter(24, FooterState{
		ModeLabel:   "COMMENT MODE",
		IsInputMode: true,
		Prompt:      "",
	}, DefaultFooterStyles()))

	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 footer lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[1], "> ") {
		t.Fatalf("expected prompt marker on line 2, got %q", lines[1])
	}
}

func TestRenderFooterInputModeShowsPromptText(t *testing.T) {
	out := stripFooterANSI(RenderFooter(40, FooterState{
		ModeLabel:   "COMMENT MODE",
		IsInputMode: true,
		Prompt:      "hello world",
	}, DefaultFooterStyles()))

	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 footer lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "> hello world") {
		t.Fatalf("expected prompt text on line 2, got %q", lines[1])
	}
}
