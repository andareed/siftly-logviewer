package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderLegendLineDoesNotTruncateEarlyWithANSI(t *testing.T) {
	line := RenderLegendLine(22, []string{"alpha", "beta"}, []lipgloss.Color{
		lipgloss.Color("#ff6b6b"),
		lipgloss.Color("#4ecdc4"),
	})

	if strings.HasSuffix(line, "...") {
		t.Fatalf("legend truncated unexpectedly: %q", line)
	}
	if w := ansi.StringWidth(line); w > 22 {
		t.Fatalf("legend width=%d exceeds max=22", w)
	}
}

func TestRenderLegendLineTruncatesToVisibleWidth(t *testing.T) {
	line := RenderLegendLine(8, []string{"alpha", "beta"}, []lipgloss.Color{
		lipgloss.Color("#ff6b6b"),
		lipgloss.Color("#4ecdc4"),
	})

	if w := ansi.StringWidth(line); w > 8 {
		t.Fatalf("legend width=%d exceeds max=8", w)
	}
}
