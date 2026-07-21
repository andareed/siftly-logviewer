package siftly

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultGraphExportNameIncludesReadableTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 21, 14, 5, 6, 0, time.UTC)
	m := Model{
		InitialPath:             filepath.Join("testdata", "today.log"),
		lastGraphExportFileName: "previous-graph.svg",
	}

	got := defaultGraphExportNameAt(m, now)
	want := "today-graph-2026-07-21_14-05-06.svg"
	if got != want {
		t.Fatalf("defaultGraphExportNameAt() = %q, want %q", got, want)
	}
}

func TestDefaultGraphExportPathUsesInputDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	now := time.Date(2026, 7, 21, 14, 5, 6, 0, time.UTC)
	m := Model{InitialPath: filepath.Join(dir, "today.log")}

	got := defaultGraphExportPathAt(m, now)
	want := filepath.Join(dir, "today-graph-2026-07-21_14-05-06.svg")
	if got != want {
		t.Fatalf("defaultGraphExportPathAt() = %q, want %q", got, want)
	}
}
