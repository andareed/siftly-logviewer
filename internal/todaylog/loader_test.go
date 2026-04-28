package todaylog

import (
	"testing"
	"time"
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
