package siftly

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestSaveModelUsesCompactRowsAndReloadsIDs(t *testing.T) {
	t.Parallel()

	m, err := NewModelFromRecords([][]string{
		{"time", "process", "value"},
		{"2026-02-11 18:30:52", "stats", "36"},
		{"2026-02-11 18:30:53", "stats", "31"},
	}, ColumnSchema{})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.table.markedRows[m.table.rows[0].ID] = ui.MarkRed
	m.table.commentRows[m.table.rows[1].ID] = "watch"
	m.table.timeWindow = featuretimewindow.Window{
		Enabled: true,
		Start:   time.Date(2026, 2, 11, 18, 30, 52, 0, time.UTC),
		End:     time.Date(2026, 2, 11, 18, 30, 53, 0, time.UTC),
	}

	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := SaveModel(m, path); err != nil {
		t.Fatalf("SaveModel: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	text := string(data)
	if strings.Contains(text, "\"cols\"") {
		t.Fatalf("compact snapshot should not repeat legacy cols keys: %s", text)
	}
	if strings.Contains(text, "\"Name\"") || strings.Contains(text, "\"Width\"") {
		t.Fatalf("compact snapshot should not persist verbose header metadata: %s", text)
	}
	if !strings.Contains(text, "\"header\":[\"time\",\"process\",\"value\"]") {
		t.Fatalf("compact snapshot should store header names once: %s", text)
	}
	if !strings.Contains(text, "\"columnLayout\"") {
		t.Fatalf("compact snapshot should store compact column layout: %s", text)
	}
	if !strings.Contains(text, "[\"2026-02-11 18:30:52\",\"stats\",\"36\"]") {
		t.Fatalf("compact snapshot should keep each row as a grep-friendly JSON array: %s", text)
	}

	var reopened Model
	if err := LoadModel(&reopened, path); err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	if got, want := len(reopened.table.rows), len(m.table.rows); got != want {
		t.Fatalf("row count = %d want %d", got, want)
	}
	for i := range m.table.rows {
		if reopened.table.rows[i].ID != m.table.rows[i].ID {
			t.Fatalf("row %d ID = %d want %d", i, reopened.table.rows[i].ID, m.table.rows[i].ID)
		}
	}
	if got := reopened.table.markedRows[m.table.rows[0].ID]; got != ui.MarkRed {
		t.Fatalf("mark = %q want %q", got, ui.MarkRed)
	}
	if got := reopened.table.commentRows[m.table.rows[1].ID]; got != "watch" {
		t.Fatalf("comment = %q want watch", got)
	}
	if got := reopened.table.header[0].Name; got != "time" {
		t.Fatalf("header[0].Name = %q want time", got)
	}
	if got := reopened.table.header[2].Role; got != ui.RoleNormal {
		t.Fatalf("header[2].Role = %d want %d", got, ui.RoleNormal)
	}
	if !reopened.table.timeWindow.Enabled {
		t.Fatalf("time window should reload as enabled")
	}
}

func TestSaveModelPreservesReorderedFrozenColumnSourceIndices(t *testing.T) {
	t.Parallel()

	m, err := NewModelFromRecords([][]string{
		{"first", "second", "third"},
		{"A", "B", "C"},
	}, ColumnSchema{})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.table.header = []ui.ColumnMeta{m.table.header[2], m.table.header[0], m.table.header[1]}
	m.table.header[0].Frozen = true
	m.table.header[1].Visible = false

	path := filepath.Join(t.TempDir(), "reordered.json")
	if err := SaveModel(m, path); err != nil {
		t.Fatalf("SaveModel: %v", err)
	}

	var reopened Model
	if err := LoadModel(&reopened, path); err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	if got := reopened.table.header[0]; got.Name != "third" || got.Index != 2 || !got.Frozen {
		t.Fatalf("reopened first column = %+v", got)
	}
	if got := reopened.table.header[1]; got.Name != "first" || got.Index != 0 || got.Visible {
		t.Fatalf("reopened second column = %+v", got)
	}
	if got := reopened.table.rows[0].Cols; len(got) != 3 || got[2] != "C" {
		t.Fatalf("source row changed on reload: %v", got)
	}
}

func TestCompactColumnLayoutAcceptsLegacyFourValueForm(t *testing.T) {
	t.Parallel()

	var layout compactColumnLayout
	if err := json.Unmarshal([]byte(`[0,true,8,1]`), &layout); err != nil {
		t.Fatalf("decode legacy column layout: %v", err)
	}
	if layout.Frozen || layout.hasIndex || layout.MinWidth != 8 || layout.Weight != 1 {
		t.Fatalf("legacy column layout = %+v", layout)
	}
}

func TestApplyColumnLayoutsRejectsDuplicateSourceIndices(t *testing.T) {
	t.Parallel()

	header := []ui.ColumnMeta{{Name: "first"}, {Name: "second"}}
	layouts := []compactColumnLayout{
		{Visible: true, MinWidth: 8, Weight: 1, Index: 0, hasIndex: true},
		{Visible: true, MinWidth: 8, Weight: 1, Index: 0, hasIndex: true},
	}
	if err := applyColumnLayouts(header, layouts); err == nil {
		t.Fatal("duplicate source indices should be rejected")
	}
}

func TestSaveModelPreservesSparseOriginalIndexes(t *testing.T) {
	t.Parallel()

	m, err := NewModelFromRecords([][]string{
		{"value"},
		{"alpha"},
		{"beta"},
		{"gamma"},
	}, ColumnSchema{})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.table.rows[1].OriginalIndex = 4
	m.table.rows[2].OriginalIndex = 5

	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := SaveModel(m, path); err != nil {
		t.Fatalf("SaveModel: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if !strings.Contains(string(data), "\"originalIndexSpans\":[[1,2]]") {
		t.Fatalf("snapshot should store original index deltas compactly: %s", data)
	}

	var reopened Model
	if err := LoadModel(&reopened, path); err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	for i, want := range []int{1, 4, 5} {
		if got := reopened.table.rows[i].OriginalIndex; got != want {
			t.Fatalf("row %d originalIndex = %d want %d", i, got, want)
		}
	}
}

func TestLoadModelSupportsLegacyRows(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "legacy.json")
	legacy := `{
  "version": 1,
  "header": [{"Name":"value","Index":0,"Role":0,"Visible":true,"MinWidth":8,"Weight":1,"Width":8}],
  "rows": [
    {"cols":["alpha"],"id":999,"originalIndex":7}
  ],
  "marked": {"999":"green"},
  "comments": {"999":"legacy note"}
}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatalf("write legacy snapshot: %v", err)
	}

	var m Model
	if err := LoadModel(&m, path); err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	if got, want := len(m.table.rows), 1; got != want {
		t.Fatalf("row count = %d want %d", got, want)
	}
	if got := m.table.rows[0].ID; got != 999 {
		t.Fatalf("legacy row ID = %d want 999", got)
	}
	if got := m.table.rows[0].OriginalIndex; got != 7 {
		t.Fatalf("legacy originalIndex = %d want 7", got)
	}
	if got := m.table.markedRows[999]; got != ui.MarkGreen {
		t.Fatalf("legacy mark = %q want %q", got, ui.MarkGreen)
	}
	if got := m.table.commentRows[999]; got != "legacy note" {
		t.Fatalf("legacy comment = %q want legacy note", got)
	}
}
