package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestRenderHeaderIsStrongSingleLineWithGutter(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	header := RenderHeader(
		3,
		[]HeaderColumn{{Name: "very long\nheading", Width: 10, Visible: true}},
		10,
		0,
		lipgloss.NewStyle().Padding(0, 1),
		lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f0f0f0")).
			Background(lipgloss.Color("#303030")),
	)
	if strings.Contains(header, "\n") {
		t.Fatalf("header should remain one line: %q", header)
	}
	if !strings.Contains(header, "│") || !strings.Contains(header, "…") {
		t.Fatalf("header missing gutter or truncation marker: %q", header)
	}
	if !strings.Contains(header, "[1;") || !strings.Contains(header, "48;2;48;48;48") {
		t.Fatalf("header missing strong emphasis: %q", header)
	}
	if width := lipgloss.Width(header); width != 13 {
		t.Fatalf("header width=%d want 13 (%q)", width, header)
	}
}
