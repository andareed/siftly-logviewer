package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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
	if len(lines) != 2 {
		t.Fatalf("expected 2 footer lines, got %d", len(lines))
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
	if len(lines) != 2 {
		t.Fatalf("expected 2 footer lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "> hello world") {
		t.Fatalf("expected prompt text on line 2, got %q", lines[1])
	}
}

func TestRenderFooterViewModeShowsStatesStatusAndHints(t *testing.T) {
	out := stripFooterANSI(RenderFooter(88, FooterState{
		ModeLabel:     "RANGE",
		ActiveStates:  []string{"FILTER", "SORT timestamp ↓", "12 SELECTED", "UNSAVED"},
		StatusMessage: "Match 4/127",
		Hints:         "j/k: extend   ctrl+c: copy   esc: clear",
	}, DefaultFooterStyles()))

	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 footer lines, got %d", len(lines))
	}
	for _, want := range []string{"RANGE", "FILTER", "SORT timestamp ↓", "12 SELECTED", "UNSAVED", "Match 4/127"} {
		if !strings.Contains(lines[0], want) {
			t.Fatalf("state line missing %q: %q", want, lines[0])
		}
	}
	if !strings.Contains(lines[1], "ctrl+c: copy") {
		t.Fatalf("action line missing contextual hints: %q", lines[1])
	}
}

func TestRenderFooterAlwaysMatchesRequestedWidth(t *testing.T) {
	for _, width := range []int{12, 24, 60} {
		out := stripFooterANSI(RenderFooter(width, FooterState{
			ModeLabel:     "FILTER",
			ActiveStates:  []string{"FILTER", "TIME WINDOW", "UNSAVED"},
			StatusMessage: "Filtering... 12s",
			Hints:         "enter: apply   ctrl+p: presets   esc: cancel",
			IsInputMode:   true,
			Prompt:        "severity=critical",
		}, DefaultFooterStyles()))
		for _, line := range strings.Split(out, "\n") {
			if got := lipgloss.Width(line); got != width {
				t.Fatalf("width %d rendered line width %d: %q", width, got, line)
			}
		}
	}
}

func TestRenderFooterDropsWholeHintsAtNarrowWidths(t *testing.T) {
	out := stripFooterANSI(RenderFooter(80, FooterState{
		ModeLabel:   "FILTER",
		IsInputMode: true,
		Prompt:      "filter: fstool_va.*seq",
		Hints:       "enter: apply   ctrl+p: presets   ctrl+h: history   esc: cancel",
	}, DefaultFooterStyles()))

	actionLine := strings.Split(out, "\n")[1]
	if !strings.Contains(actionLine, "enter: apply") {
		t.Fatalf("expected highest-priority complete hint: %q", actionLine)
	}
	if strings.HasSuffix(strings.TrimSpace(actionLine), "ctrl+") {
		t.Fatalf("hint should not be truncated mid-shortcut: %q", actionLine)
	}
}
