package dialogs

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestCommandPaletteFiltersAcrossCommandMetadata(t *testing.T) {
	d := NewCommandPalette([]CommandItem{
		{ID: "output.save", Category: "Output", Title: "Save Siftly JSON", Shortcut: "s", Enabled: true},
		{ID: "output.graph", Category: "Output", Title: "Export graph SVG", Shortcut: "e g", Keywords: "chart plot", Enabled: true},
	}, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))
	d.Show()

	for _, r := range "chart" {
		_, _, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if len(d.filtered) != 1 || d.filtered[0].ID != "output.graph" {
		t.Fatalf("filtered commands = %#v, want graph export", d.filtered)
	}
	view := d.View()
	if !strings.Contains(view, "Export graph SVG") || strings.Contains(view, "Save Siftly JSON") {
		t.Fatalf("palette view does not reflect search: %q", view)
	}
}

func TestCommandPaletteRanksCategoryAndTitleMatchesFirst(t *testing.T) {
	d := NewCommandPalette([]CommandItem{
		{ID: "navigation.left", Category: "Navigation", Title: "Scroll columns left", Description: "Move the table viewport", Shortcut: "h", Enabled: true},
		{ID: "view.columns", Category: "View and Columns", Title: "Choose visible columns", Shortcut: "v c", Enabled: true},
	}, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))
	d.Show()

	_, _, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("columns")})
	if len(d.filtered) != 2 || d.filtered[0].ID != "view.columns" {
		t.Fatalf("ranked commands = %#v, want view columns first", d.filtered)
	}
	if view := d.View(); !strings.Contains(view, "VIEW AND COLUMNS") {
		t.Fatalf("palette clips the command category: %q", view)
	}
}

func TestCommandPaletteRunsSelectedEnabledCommand(t *testing.T) {
	d := NewCommandPalette([]CommandItem{
		{ID: "history.undo", Category: "History", Title: "Undo", Shortcut: "u", Enabled: false, DisabledReason: "Nothing to undo"},
		{ID: "filter.open", Category: "Search", Title: "Filter rows", Shortcut: "f", Enabled: true},
	}, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))

	_, action, _ := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if action.Kind != ActionCommandRun || action.CommandID != "filter.open" {
		t.Fatalf("enter action = %+v, want enabled filter command", action)
	}
}

func TestCommandPaletteDoesNotRunDisabledCommand(t *testing.T) {
	d := NewCommandPalette([]CommandItem{{
		ID: "history.undo", Category: "History", Title: "Undo", Shortcut: "u", Enabled: false, DisabledReason: "Nothing to undo",
	}}, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))

	_, action, _ := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if action.Kind != ActionNone {
		t.Fatalf("disabled command action = %+v, want no action", action)
	}
	if !strings.Contains(d.View(), "Nothing to undo") {
		t.Fatalf("disabled reason missing from palette: %q", d.View())
	}
}

func TestHelpGroupsCommandsByCategory(t *testing.T) {
	d := NewHelpDialog([]CommandItem{
		{ID: "general.quit", Category: "General", Title: "Quit", Shortcut: "q", Enabled: true},
		{ID: "navigation.down", Category: "Navigation", Title: "Move down", Shortcut: "j", Enabled: true},
	}, 100, 30)

	view := d.View()
	for _, want := range []string{"GENERAL", "Quit", "NAVIGATION", "Move down"} {
		if !strings.Contains(view, want) {
			t.Fatalf("categorized help missing %q: %q", want, view)
		}
	}
}
