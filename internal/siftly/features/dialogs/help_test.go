package dialogs

import (
	"strings"
	"testing"
)

func TestHelpViewUsesActionRowInsteadOfInstructionFooter(t *testing.T) {
	d := NewHelpDialog([]CommandItem{{
		ID:          "general.quit",
		Category:    "General",
		Title:       "Quit",
		Shortcut:    "q",
		Description: "Exit Siftly after confirming whether unsaved changes should be discarded",
		Enabled:     true,
	}}, 100, 30)

	v := d.View()
	if strings.Contains(v, "enter/esc to return") || strings.Contains(v, "esc: close") {
		t.Fatalf("help dialog still contains legacy instruction text: %q", v)
	}
	if !strings.Contains(v, "[ Esc Close ]") {
		t.Fatalf("help dialog missing action row: %q", v)
	}
	if !strings.Contains(v, "Keyboard Reference") || !strings.Contains(v, "GENERAL") || !strings.Contains(v, "Quit") {
		t.Fatalf("help dialog is not categorized: %q", v)
	}
	for _, want := range []string{"KEY", "ACTION", "DETAIL", "Exit Siftly after"} {
		if !strings.Contains(v, want) {
			t.Fatalf("help dialog missing %q: %q", want, v)
		}
	}
}

func TestHelpViewCollapsesDetailColumnOnNarrowTerminals(t *testing.T) {
	d := NewHelpDialog([]CommandItem{{
		Category:    "Output and Data",
		Title:       "Save Siftly JSON",
		Shortcut:    "s",
		Description: "Write a reloadable Siftly file",
		Enabled:     true,
	}}, 52, 16)

	view := d.View()
	if strings.Contains(view, "DETAIL") || strings.Contains(view, "Write a reloadable") {
		t.Fatalf("compact help should omit the detail column: %q", view)
	}
	if !strings.Contains(view, "Save Siftly JSON") {
		t.Fatalf("compact help omitted the command action: %q", view)
	}
}

func TestHelpStatusCountsCommandsInsteadOfLayoutLines(t *testing.T) {
	d := NewHelpDialog([]CommandItem{
		{Category: "General", Title: "Quit", Shortcut: "q", Enabled: true},
		{Category: "Navigation", Title: "Move down", Shortcut: "j", Enabled: true},
	}, 100, 30)

	if view := d.View(); !strings.Contains(view, "Commands 1-2 of 2") {
		t.Fatalf("help status does not report command positions: %q", view)
	}
}
