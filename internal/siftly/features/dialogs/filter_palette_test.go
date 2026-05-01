package dialogs

import "testing"

func TestNewFilterHistoryPaletteDialogStartsOnHistory(t *testing.T) {
	lastFilterPaletteTab = filterTabPresets

	d := NewFilterHistoryPaletteDialog(
		[]FilterPreset{{Pattern: "preset-pattern", Description: "preset"}},
		[]string{"history-pattern"},
		100,
		30,
		"",
		"",
	)

	if d.activeTab != filterTabHistory {
		t.Fatalf("active tab = %q, want %q", d.activeTab, filterTabHistory)
	}
	pattern, ok := d.selectedPattern()
	if !ok {
		t.Fatalf("expected selected history pattern")
	}
	if pattern != "history-pattern" {
		t.Fatalf("selected pattern = %q, want %q", pattern, "history-pattern")
	}
	if lastFilterPaletteTab != filterTabPresets {
		t.Fatalf("history constructor should not mutate last tab, got %q", lastFilterPaletteTab)
	}
}
