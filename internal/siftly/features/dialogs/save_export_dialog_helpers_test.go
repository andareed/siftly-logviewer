package dialogs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveFileDialogStateNewFile(t *testing.T) {
	dir := t.TempDir()
	st := resolveFileDialogState(dir, "new.json", "", "Save")

	if st.TopRightState != "" {
		t.Fatalf("expected empty top-right state for new file, got %q", st.TopRightState)
	}
	if st.StatusKind != "success" {
		t.Fatalf("expected success status for new file, got %q", st.StatusKind)
	}
	if !st.PrimaryEnabled || st.PrimaryAction != "Save" {
		t.Fatalf("expected enabled Save action, got enabled=%t action=%q", st.PrimaryEnabled, st.PrimaryAction)
	}
}

func TestResolveFileDialogStateOverwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "exists.json")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	st := resolveFileDialogState(dir, "exists.json", "", "Save")
	if st.TopRightState != "OVERWRITE" {
		t.Fatalf("expected overwrite state, got %q", st.TopRightState)
	}
	if st.StatusKind != "warn" {
		t.Fatalf("expected warn status, got %q", st.StatusKind)
	}
	if !st.PrimaryEnabled || st.PrimaryAction != "Overwrite" {
		t.Fatalf("expected enabled Overwrite action, got enabled=%t action=%q", st.PrimaryEnabled, st.PrimaryAction)
	}
}

func TestResolveFileDialogStateInvalid(t *testing.T) {
	st := resolveFileDialogState("", "", "", "Save")
	if st.TopRightState != "INVALID" {
		t.Fatalf("expected invalid state, got %q", st.TopRightState)
	}
	if st.StatusKind != "error" {
		t.Fatalf("expected error status, got %q", st.StatusKind)
	}
	if st.PrimaryEnabled {
		t.Fatalf("expected disabled primary action for invalid state")
	}
}

func TestSaveAndExportViewsDoNotContainLegacyInstructionHints(t *testing.T) {
	dir := t.TempDir()
	save := NewSaveDialog("hostlog.json", dir)
	export := NewExportDialog("hostlog.csv", dir)

	saveView := save.View()
	exportView := export.View()

	for _, v := range []string{saveView, exportView} {
		if strings.Contains(v, "enter to save") || strings.Contains(v, "enter: save") {
			t.Fatalf("view contains legacy save hint: %q", v)
		}
		if strings.Contains(v, "enter to export") || strings.Contains(v, "enter: export") {
			t.Fatalf("view contains legacy export hint: %q", v)
		}
	}
}
