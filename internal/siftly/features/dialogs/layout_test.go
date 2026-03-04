package dialogs

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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
