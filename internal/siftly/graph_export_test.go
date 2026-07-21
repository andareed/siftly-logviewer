package siftly

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	if !strings.Contains(text, "metric.a") {
		t.Fatalf("export should include filtered series: %s", text)
	}
	if strings.Contains(text, "metric.b") {
		t.Fatalf("export should not include filtered-out series: %s", text)
	}
}
