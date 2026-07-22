package siftly

import (
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

func TestCommandCatalogHasUniqueExecutableEntries(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.graphConfig.Enabled = true
	m.SetFullSourceReload(func() (*Model, error) { return m, nil })

	seen := make(map[string]struct{})
	catalog := m.commandCatalog()
	if len(catalog) < 45 {
		t.Fatalf("command catalog has only %d entries", len(catalog))
	}
	for _, command := range catalog {
		item := command.item
		if item.ID == "" || item.Category == "" || item.Title == "" || item.Shortcut == "" {
			t.Fatalf("incomplete command item: %+v", item)
		}
		if len(command.sequence) == 0 {
			t.Fatalf("command %q has no executable key sequence", item.ID)
		}
		if _, exists := seen[item.ID]; exists {
			t.Fatalf("duplicate command ID %q", item.ID)
		}
		seen[item.ID] = struct{}{}
	}

	for _, id := range []string{
		commandPaletteOpen,
		commandHelpOpen,
		commandQuit,
		commandUndo,
		commandRedo,
		commandFilter,
		commandMarkOpen,
		commandCommentEdit,
		commandColumns,
		commandTimeWindow,
		commandGraphToggle,
		commandSave,
		commandExportData,
		commandExportGraph,
		commandReloadFullSource,
	} {
		if _, exists := seen[id]; !exists {
			t.Fatalf("command catalog omits %q", id)
		}
	}
}

func TestCommandCatalogUsesExistingPrefixSequences(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.graphConfig.Enabled = true

	tests := map[string][]string{
		commandCommentEdit:   {"c", "e"},
		commandColumns:       {"v"},
		commandTimeWindow:    {"t", "w"},
		commandExportData:    {"e", "d"},
		commandExportGraph:   {"e", "g"},
		commandFilterPresets: {"f", "ctrl+p"},
	}

	for _, command := range m.commandCatalog() {
		want, ok := tests[command.item.ID]
		if !ok {
			continue
		}
		if len(command.sequence) != len(want) {
			t.Fatalf("%s sequence length = %d, want %d", command.item.ID, len(command.sequence), len(want))
		}
		for i, keyMsg := range command.sequence {
			if keyMsg.String() != want[i] {
				t.Fatalf("%s key %d = %q, want %q", command.item.ID, i, keyMsg.String(), want[i])
			}
		}
		delete(tests, command.item.ID)
	}
	for id := range tests {
		t.Fatalf("command catalog omits sequence test for %q", id)
	}
}

func TestCommandCatalogHidesUnsupportedFeatureCommands(t *testing.T) {
	m := newChangeTrackingTestModel()
	for _, item := range m.commandItems() {
		if item.ID == commandGraphToggle || item.ID == commandExportGraph || item.ID == commandReloadFullSource {
			t.Fatalf("unsupported command %q should be hidden", item.ID)
		}
	}
}

func TestCommandPaletteUsesPlainP(t *testing.T) {
	m := newChangeTrackingTestModel()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	if got := m.resolveViewAction(msg); got != viewActionOpenPalette {
		t.Fatalf("p action = %v, want command palette", got)
	}
}

func TestPaletteCommandExecutesExistingPrefixHandler(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.runPaletteCommand(commandColumns)

	if _, ok := m.activeDialog.(*dialogs.ColumnManager); !ok {
		t.Fatalf("palette columns command opened %T, want column manager", m.activeDialog)
	}
}

func TestCommandPaletteSelectionExecutesThroughModelUpdate(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.ready = true
	m.terminalWidth = 100
	m.terminalHeight = 30
	m.openCommandPalette()

	for _, value := range "columns" {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{value}})
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if _, ok := m.activeDialog.(*dialogs.ColumnManager); !ok {
		t.Fatalf("palette model update opened %T, want column manager", m.activeDialog)
	}
}
