package dialogs

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestHelpViewUsesActionRowInsteadOfInstructionFooter(t *testing.T) {
	d := NewHelpDialog([]key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	})

	v := d.View()
	if strings.Contains(v, "enter/esc to return") || strings.Contains(v, "esc: close") {
		t.Fatalf("help dialog still contains legacy instruction text: %q", v)
	}
	if !strings.Contains(v, "[ Esc Close ]") {
		t.Fatalf("help dialog missing action row: %q", v)
	}
}
