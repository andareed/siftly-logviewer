package dialogs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDialogsFitCompactTerminalAndResizeWhileOpen(t *testing.T) {
	const terminalWidth = 52
	const terminalHeight = 16
	commands := []CommandItem{
		{Category: "General", Title: "Open command palette", Shortcut: "p", Enabled: true},
		{Category: "Navigation", Title: "Move down", Shortcut: "j", Enabled: true},
		{Category: "Output", Title: "Export graph SVG", Shortcut: "e g", Enabled: true},
	}
	columns := []ColumnManagerItem{
		{Name: "timestamp", Visible: true},
		{Name: "diagnostic message", Visible: true},
		{Name: "value", Visible: true},
	}

	dialogs := []Dialog{
		NewCommandPalette(commands, 120, 40, "", ""),
		NewHelpDialog(commands, 120, 40),
		NewColumnManager(columns, columns, false, -1, false, 120, 40, "", ""),
		NewFilterPaletteDialog(
			[]FilterPreset{{Pattern: "severity=error", Description: "Errors"}},
			[]string{"host=example-host"},
			120,
			40,
			"",
			"",
		),
	}

	for _, dialog := range dialogs {
		resizable, ok := dialog.(Resizable)
		if !ok {
			t.Fatalf("%T does not support responsive resizing", dialog)
		}
		resizable.Resize(terminalWidth, terminalHeight)
		view := dialog.View()
		if width := lipgloss.Width(view); width > terminalWidth {
			t.Fatalf("%T width = %d, terminal width %d", dialog, width, terminalWidth)
		}
		if height := lipgloss.Height(view); height > terminalHeight {
			t.Fatalf("%T height = %d, terminal height %d", dialog, height, terminalHeight)
		}
	}
}

func TestCommandPaletteUsesOnlyRowsNeededByContent(t *testing.T) {
	dialog := NewCommandPalette([]CommandItem{
		{Category: "General", Title: "Quit", Shortcut: "q", Enabled: true},
		{Category: "General", Title: "Help", Shortcut: "?", Enabled: true},
	}, 140, 50, "", "")

	if got, want := lipgloss.Height(dialog.View()), 11; got != want {
		t.Fatalf("two-command palette height = %d, want content height %d", got, want)
	}
}

func TestFileDialogsTrimPreviewToTerminalHeight(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 20; i++ {
		name := filepath.Join(dir, fmt.Sprintf("entry-%02d", i))
		if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
	}

	dialogs := []Dialog{
		NewSaveDialog("output.json", dir),
		NewExportDialog("output.csv", dir),
	}
	for _, dialog := range dialogs {
		dialog.(Resizable).Resize(56, 16)
		view := dialog.View()
		if width := lipgloss.Width(view); width > 56 {
			t.Fatalf("%T width = %d, terminal width 56", dialog, width)
		}
		if height := lipgloss.Height(view); height > 16 {
			t.Fatalf("%T height = %d, terminal height 16", dialog, height)
		}
	}
}
