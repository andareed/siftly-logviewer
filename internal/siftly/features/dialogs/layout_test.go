package dialogs

import (
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestRenderDialogPanelFlattensMultilineRows(t *testing.T) {
	const width = 40
	out := renderDialogPanel("Save", "enter: save", width, []string{
		"line1",
		"file-a\nfile-b\nfile-c",
		"line2",
	})

	lines := strings.Split(out, "\n")
	if len(lines) != 7 {
		t.Fatalf("unexpected line count: got %d want 7", len(lines))
	}
	for i, line := range lines {
		if lipgloss.Width(line) != width {
			t.Fatalf("line %d width mismatch: got %d want %d", i, lipgloss.Width(line), width)
		}
	}
}

func TestRenderDialogPanelUsesSharedBorderTokens(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(previousProfile)

	tokens := ui.DefaultDesignTokens()
	out := renderDialogPanel("Save", "ready", 40, []string{"content"}, tokens)
	if !strings.Contains(out, "38;2;138;138;138") {
		t.Fatalf("dialog does not use the strong shared border colour: %q", out)
	}
	if !strings.Contains(out, "38;2;240;240;240") {
		t.Fatalf("dialog does not use the strong shared title emphasis: %q", out)
	}
}
