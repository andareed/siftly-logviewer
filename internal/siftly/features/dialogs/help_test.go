package dialogs

import (
	"strings"
	"testing"
)

func TestHelpViewUsesActionRowInsteadOfInstructionFooter(t *testing.T) {
	d := NewHelpDialog([]CommandItem{{
		ID:       "general.quit",
		Category: "General",
		Title:    "Quit",
		Shortcut: "q",
		Enabled:  true,
	}}, 100, 30)

	v := d.View()
	if strings.Contains(v, "enter/esc to return") || strings.Contains(v, "esc: close") {
		t.Fatalf("help dialog still contains legacy instruction text: %q", v)
	}
	if !strings.Contains(v, "[ Esc Close ]") {
		t.Fatalf("help dialog missing action row: %q", v)
	}
	if !strings.Contains(v, "GENERAL") || !strings.Contains(v, "Quit") {
		t.Fatalf("help dialog is not categorized: %q", v)
	}
}
