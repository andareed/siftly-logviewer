package siftly

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadFilterConfigMergesDefaultsPresetsAndHistory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	presetsPath := filepath.Join(dir, "presets.json")
	historyPath := filepath.Join(dir, "history.json")

	if err := os.WriteFile(presetsPath, []byte(`{
  "presets": [
    {"pattern": "override", "description": "from file"},
    {"pattern": "file-only", "description": "extra file preset"}
  ]
}
`), 0o644); err != nil {
		t.Fatalf("write presets: %v", err)
	}
	if err := os.WriteFile(historyPath, []byte(`{
  "history": [" recent ", "other", "recent", ""]
}
`), 0o644); err != nil {
		t.Fatalf("write history: %v", err)
	}

	m := &Model{}
	m.SetFilterConfig(FilterConfigSettings{
		DefaultPresets: PresetList{
			{Pattern: "base", Description: "built-in"},
			{Pattern: "override", Description: "built-in value"},
		},
		PresetsPath: presetsPath,
		HistoryPath: historyPath,
	})

	cfg, err := m.loadFilterConfig()
	if err != nil {
		t.Fatalf("loadFilterConfig: %v", err)
	}

	wantPresets := PresetList{
		{Pattern: "base", Description: "built-in"},
		{Pattern: "override", Description: "from file"},
		{Pattern: "file-only", Description: "extra file preset"},
	}
	if !reflect.DeepEqual(cfg.Presets, wantPresets) {
		t.Fatalf("presets mismatch\n got: %#v\nwant: %#v", cfg.Presets, wantPresets)
	}

	wantHistory := []string{"recent", "other"}
	if !reflect.DeepEqual(cfg.History, wantHistory) {
		t.Fatalf("history mismatch\n got: %#v\nwant: %#v", cfg.History, wantHistory)
	}
}

func TestRecordFilterHistoryUsesSeparateHistoryFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	presetsPath := filepath.Join(dir, "hostlog-filters.json")
	historyPath := filepath.Join(dir, "state", "hostlog-filter-history.json")

	presetBody := `{
  "presets": [
    {"pattern": "built-in", "description": "do not rewrite"}
  ]
}
`
	if err := os.WriteFile(presetsPath, []byte(presetBody), 0o644); err != nil {
		t.Fatalf("write presets: %v", err)
	}

	m := &Model{}
	m.SetFilterConfig(FilterConfigSettings{
		PresetsPath: presetsPath,
		HistoryPath: historyPath,
	})

	if err := m.recordFilterHistory("  first  "); err != nil {
		t.Fatalf("record first: %v", err)
	}
	if err := m.recordFilterHistory("second"); err != nil {
		t.Fatalf("record second: %v", err)
	}
	if err := m.recordFilterHistory("first"); err != nil {
		t.Fatalf("record duplicate: %v", err)
	}

	historyCfg, err := LoadFilterHistoryConfig(historyPath)
	if err != nil {
		t.Fatalf("LoadFilterHistoryConfig: %v", err)
	}
	wantHistory := []string{"first", "second"}
	if !reflect.DeepEqual(historyCfg.History, wantHistory) {
		t.Fatalf("history mismatch\n got: %#v\nwant: %#v", historyCfg.History, wantHistory)
	}

	presetBytes, err := os.ReadFile(presetsPath)
	if err != nil {
		t.Fatalf("read presets after history write: %v", err)
	}
	if string(presetBytes) != presetBody {
		t.Fatalf("preset file changed unexpectedly\n got: %s\nwant: %s", string(presetBytes), presetBody)
	}
}
