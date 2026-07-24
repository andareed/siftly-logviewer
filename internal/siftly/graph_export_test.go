package siftly

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestGraphColumnIndicesFollowStableSourcesAfterReorder(t *testing.T) {
	m, err := NewModelFromRecords([][]string{
		{"timestamp", "key", "value"},
		{"1713878400", "metric.a", "10"},
	}, ColumnSchema{})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.SetGraphConfig(GraphConfig{
		Enabled:      true,
		TimeColumn:   "timestamp",
		SeriesColumn: "key",
		ValueColumn:  "value",
	})
	m.table.header = []ui.ColumnMeta{m.table.header[2], m.table.header[0], m.table.header[1]}

	timeColumn, seriesColumn, valueColumn, ok := m.graphColumnIndices()
	if !ok || timeColumn != 0 || seriesColumn != 1 || valueColumn != 2 {
		t.Fatalf("graph source columns = (%d,%d,%d,%t)", timeColumn, seriesColumn, valueColumn, ok)
	}
}

func TestExportGraphModelWritesSVGForFilteredRows(t *testing.T) {
	t.Parallel()

	m, err := NewModelFromRecords([][]string{
		{"timestamp", "key", "value"},
		{"1713878400", "metric.a", "10"},
		{"1713878460", "metric.b", "20"},
		{"1713878520", "metric.a", "30"},
	}, ColumnSchema{})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.InitialPath = filepath.Join(t.TempDir(), "today.log")
	m.SetGraphConfig(GraphConfig{
		Enabled:      true,
		TimeColumn:   "timestamp",
		SeriesColumn: "key",
		ValueColumn:  "value",
		MaxKeys:      8,
		ScaleMode:    "linear",
		Aggregate:    "last",
		FillMode:     "none",
	})
	m.InitialiseView()
	if err := m.setFilterPattern(`metric\.a`); err != nil {
		t.Fatalf("set filter: %v", err)
	}

	path := filepath.Join(t.TempDir(), "today-graph.svg")
	if err := ExportGraphModel(m, path); err != nil {
		t.Fatalf("ExportGraphModel: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read svg: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `<svg`) {
		t.Fatalf("export should be svg: %s", text)
	}
	if !strings.Contains(text, `width="1920" height="1080" viewBox="0 0 1920 1080"`) {
		t.Fatalf("export should use the full-HD default canvas: %s", text)
	}
	if !strings.Contains(text, "metric.a") {
		t.Fatalf("export should include filtered series: %s", text)
	}
	if strings.Contains(text, "metric.b") {
		t.Fatalf("export should not include filtered-out series: %s", text)
	}
}
