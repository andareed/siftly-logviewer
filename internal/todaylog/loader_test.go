package todaylog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andareed/siftly-hostlog/internal/siftly"
)

func TestParseStatsLine(t *testing.T) {
	t.Parallel()

	record, count, err := parseStatsLine("0 1713878400 proc-name 123 metric_name some long value here")
	if err != nil {
		t.Fatalf("parseStatsLine returned error: %v", err)
	}
	if count != 6 {
		t.Fatalf("field count = %d want 6", count)
	}

	want := []string{
		time.Unix(1713878400, 0).Format("2006-01-02 15:04:05"),
		"1713878400",
		"123",
		"proc-name",
		"metric_name",
		"some long value here",
	}
	if len(record) != len(want) {
		t.Fatalf("record len = %d want %d", len(record), len(want))
	}
	for i := range want {
		if record[i] != want[i] {
			t.Fatalf("record[%d] = %q want %q", i, record[i], want[i])
		}
	}
}

func TestParseStatsLineReportsShortInput(t *testing.T) {
	t.Parallel()

	record, count, err := parseStatsLine("0 1713878400 proc-name 123 metric_name")
	if err != nil {
		t.Fatalf("parseStatsLine returned error: %v", err)
	}
	if record != nil {
		t.Fatalf("record = %#v want nil", record)
	}
	if count != 5 {
		t.Fatalf("field count = %d want 5", count)
	}
}

func TestLoadModelAutoLoadsSavedJSONSnapshot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "today.log")
	if err := os.WriteFile(logPath, []byte("0 1713878400 proc-name 123 metric_name some long value here\n"), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}

	original, err := LoadModelAuto(logPath)
	if err != nil {
		t.Fatalf("load source log: %v", err)
	}

	snapshotPath := filepath.Join(dir, "today.json")
	if err := siftly.SaveModel(original, snapshotPath); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}

	reopened, err := LoadModelAuto(snapshotPath)
	if err != nil {
		t.Fatalf("reopen snapshot: %v", err)
	}
	if reopened.InitialPath != snapshotPath {
		t.Fatalf("InitialPath = %q want %q", reopened.InitialPath, snapshotPath)
	}

	resavedPath := filepath.Join(dir, "today-roundtrip.json")
	if err := siftly.SaveModel(reopened, resavedPath); err != nil {
		t.Fatalf("resave snapshot: %v", err)
	}

	data, err := os.ReadFile(resavedPath)
	if err != nil {
		t.Fatalf("read resaved snapshot: %v", err)
	}

	var snapshot struct {
		Version int        `json:"version"`
		Rows    [][]string `json:"rows"`
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("unmarshal resaved snapshot: %v", err)
	}

	if snapshot.Version != 2 {
		t.Fatalf("snapshot version = %d want 2", snapshot.Version)
	}
	if len(snapshot.Rows) != 1 {
		t.Fatalf("row count = %d want 1", len(snapshot.Rows))
	}
	if got, want := len(snapshot.Rows[0]), 6; got != want {
		t.Fatalf("column count = %d want %d", got, want)
	}
	if got, want := snapshot.Rows[0][3], "proc-name"; got != want {
		t.Fatalf("process column = %q want %q", got, want)
	}
	if got, want := snapshot.Rows[0][5], "some long value here"; got != want {
		t.Fatalf("value column = %q want %q", got, want)
	}
	if !strings.Contains(string(data), "\"rows\"") {
		t.Fatalf("resaved snapshot missing rows payload")
	}
	if strings.Contains(string(data), "\"cols\"") {
		t.Fatalf("resaved snapshot should use compact row arrays, got legacy cols keys")
	}
}
