package dialogs

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestRecoveryDialogRequiresAnExplicitDecision(t *testing.T) {
	dialog := NewRecoveryDialog("today.log", "24 Jul 2026 09:23:56 BST", "2 marks", 80, 24)

	for _, key := range []string{"enter", "esc", "x"} {
		_, action, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		if action.Kind != ActionNone {
			t.Fatalf("%q action = %v, want no action", key, action.Kind)
		}
	}

	tests := []struct {
		key  rune
		want ActionKind
	}{
		{key: 'r', want: ActionRecoveryRestore},
		{key: 'd', want: ActionRecoveryDiscard},
		{key: 'q', want: ActionRecoveryQuit},
	}
	for _, tt := range tests {
		_, action, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
		if action.Kind != tt.want {
			t.Fatalf("%q action = %v, want %v", tt.key, action.Kind, tt.want)
		}
	}
}

func TestRecoveryDialogExplainsThePendingStateAndFitsTerminal(t *testing.T) {
	dialog := NewRecoveryDialog(
		"today.log",
		"24 Jul 2026 09:23:56 BST",
		"2 marks, 1 comment",
		52,
		16,
	)
	view := dialog.View()

	for _, want := range []string{"Recovery Found", "today.log", "2 marks, 1 comment", "r Restore", "d Discard"} {
		if !strings.Contains(view, want) {
			t.Fatalf("recovery dialog does not contain %q: %q", want, view)
		}
	}
	if width := lipgloss.Width(view); width > 52 {
		t.Fatalf("recovery dialog width = %d, terminal width 52", width)
	}
	if height := lipgloss.Height(view); height > 16 {
		t.Fatalf("recovery dialog height = %d, terminal height 16", height)
	}
}
