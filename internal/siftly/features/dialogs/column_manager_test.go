package dialogs

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestColumnManagerFiltersAndAppliesDraft(t *testing.T) {
	dialog := NewColumnManager([]ColumnManagerItem{
		{SourceIndex: 0, Name: "timestamp", Visible: true, MinWidth: 12},
		{SourceIndex: 1, Name: "message", Visible: true, MinWidth: 20},
	}, nil, false, -1, false, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))

	_, _, _ = dialog.Update(keyRunes('/'))
	for _, value := range "message" {
		_, _, _ = dialog.Update(keyRunes(value))
	}
	_, _, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, _, _ = dialog.Update(keyRunes(' '))
	_, _, _ = dialog.Update(keyRunes('s'))
	_, _, _ = dialog.Update(keyRunes('a'))
	_, action, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if action.Kind != ActionColumnManagerApply || action.ColumnManager == nil {
		t.Fatalf("apply action = %+v", action)
	}
	result := action.ColumnManager
	if result.Columns[1].Visible {
		t.Fatalf("message column should be hidden in result: %+v", result.Columns[1])
	}
	if !result.Columns[1].AutoFit {
		t.Fatalf("message column should have auto-fit queued")
	}
	if !result.SortEnabled || result.SortColumn != 1 || result.SortDesc {
		t.Fatalf("sort result = enabled:%t column:%d desc:%t", result.SortEnabled, result.SortColumn, result.SortDesc)
	}
}

func TestColumnManagerFreezeKeepsPinnedColumnsFirst(t *testing.T) {
	dialog := NewColumnManager([]ColumnManagerItem{
		{SourceIndex: 0, Name: "first", Visible: true},
		{SourceIndex: 1, Name: "second", Visible: true},
		{SourceIndex: 2, Name: "third", Visible: true},
	}, nil, false, -1, false, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))

	_, _, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, _, _ = dialog.Update(keyRunes('f'))
	_, action, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})

	result := action.ColumnManager
	if result == nil || result.Columns[0].SourceIndex != 1 || !result.Columns[0].Frozen {
		t.Fatalf("frozen column was not pinned first: %+v", result)
	}
}

func TestColumnManagerReordersWithJK(t *testing.T) {
	dialog := NewColumnManager([]ColumnManagerItem{
		{SourceIndex: 0, Name: "first", Visible: true},
		{SourceIndex: 1, Name: "second", Visible: true},
	}, nil, false, -1, false, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))

	_, _, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyDown})
	_, _, _ = dialog.Update(keyRunes('K'))
	_, action, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if action.ColumnManager == nil || action.ColumnManager.Columns[0].SourceIndex != 1 {
		t.Fatalf("reordered result = %+v", action.ColumnManager)
	}
}

func TestColumnManagerCancelDiscardsDraft(t *testing.T) {
	dialog := NewColumnManager([]ColumnManagerItem{
		{SourceIndex: 0, Name: "first", Visible: true},
	}, nil, false, -1, false, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))
	_, _, _ = dialog.Update(keyRunes(' '))
	_, action, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if action.Kind != ActionColumnManagerCancel || action.ColumnManager != nil {
		t.Fatalf("cancel action = %+v", action)
	}
}

func TestColumnManagerResetRestoresDraftDefaults(t *testing.T) {
	current := []ColumnManagerItem{
		{SourceIndex: 2, Name: "third", Visible: false, AutoFit: true},
		{SourceIndex: 0, Name: "first", Visible: true},
		{SourceIndex: 1, Name: "second", Visible: true, Frozen: true},
	}
	defaults := []ColumnManagerItem{
		{SourceIndex: 0, Name: "first", Visible: true, MinWidth: 8, Weight: 1},
		{SourceIndex: 1, Name: "second", Visible: true, MinWidth: 12, Weight: 2},
		{SourceIndex: 2, Name: "third", Visible: false, MinWidth: 8},
	}
	dialog := NewColumnManager(current, defaults, true, 1, true, 100, 30, lipgloss.Color("15"), lipgloss.Color("8"))

	_, _, _ = dialog.Update(keyRunes('r'))
	if view := dialog.View(); !strings.Contains(view, "Layout reset queued") || !strings.Contains(view, "r reset") {
		t.Fatalf("reset state is not visible in manager: %q", view)
	}
	_, action, _ := dialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if action.Kind != ActionColumnManagerApply || action.ColumnManager == nil {
		t.Fatalf("reset apply action = %+v", action)
	}
	result := action.ColumnManager
	if result.SortEnabled || result.SortColumn != -1 || result.SortDesc {
		t.Fatalf("reset retained sort: %+v", result)
	}
	for i, column := range result.Columns {
		if column.SourceIndex != i || column.Visible != defaults[i].Visible || column.Frozen || column.AutoFit ||
			column.MinWidth != defaults[i].MinWidth || column.Weight != defaults[i].Weight {
			t.Fatalf("reset column %d = %+v", i, column)
		}
	}
}

func keyRunes(value rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{value}}
}
