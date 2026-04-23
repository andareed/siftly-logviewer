package siftly

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func TestMetaStatusViewNoFilterNoMarks(t *testing.T) {
	m := Model{
		fileName: "hostlog.json",
		cursor:   5,
		table: tableState{
			filteredIndices: make([]int, 300),
		},
	}

	got := stripANSI(m.metaStatusView(60))
	if lipgloss.Width(got) != 60 {
		t.Fatalf("header width mismatch: got %d want 60 (%q)", lipgloss.Width(got), got)
	}
	if strings.Contains(got, "Filter:") {
		t.Fatalf("filter field should be omitted when inactive: %q", got)
	}
	if strings.Contains(got, "Marks:") {
		t.Fatalf("marks field should be omitted when off: %q", got)
	}
	if !strings.HasSuffix(got, "Rows 6/300") {
		t.Fatalf("rows field missing or misaligned: %q", got)
	}
}

func TestMetaStatusViewWithFilter(t *testing.T) {
	m := Model{
		fileName: "hostlog.json",
		cursor:   5,
		table: tableState{
			filterPattern:   "severity=error",
			filterEnabled:   true,
			filterRegex:     mustCompileFilter(t, `severity=error`),
			filteredIndices: make([]int, 300),
		},
	}

	got := stripANSI(m.metaStatusView(90))
	wantSuffix := "Rows 6/300  Filter: severity=error (on)"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("right block mismatch: got %q want suffix %q", got, wantSuffix)
	}
	if strings.Contains(got, "Marks:") {
		t.Fatalf("marks field should be omitted when off: %q", got)
	}
}

func TestMetaStatusViewWithFilterAndMarks(t *testing.T) {
	m := Model{
		fileName: "hostlog.json",
		cursor:   5,
		table: tableState{
			showOnlyMarked:  true,
			filterPattern:   "severity=error",
			filterEnabled:   true,
			filterRegex:     mustCompileFilter(t, `severity=error`),
			filteredIndices: make([]int, 300),
		},
	}

	got := stripANSI(m.metaStatusView(110))
	wantSuffix := "Rows 6/300  Filter: severity=error (on)  Marks: on"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("right block mismatch: got %q want suffix %q", got, wantSuffix)
	}
}

func TestMetaStatusViewWithDisabledFilter(t *testing.T) {
	m := Model{
		fileName: "hostlog.json",
		cursor:   5,
		table: tableState{
			filterPattern:   "severity=error",
			filterEnabled:   false,
			filterRegex:     mustCompileFilter(t, `severity=error`),
			filteredIndices: make([]int, 300),
		},
	}

	got := stripANSI(m.metaStatusView(82))
	wantSuffix := "Rows 6/300  Filter: severity=error (off)"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("right block mismatch: got %q want suffix %q", got, wantSuffix)
	}
}

func TestTruncateFilenameMiddlePreserveExt(t *testing.T) {
	got := truncateFilenameMiddlePreserveExt("very_long_file_name.log", 16)
	want := "very_lo…name.log"
	if got != want {
		t.Fatalf("unexpected truncation: got %q want %q", got, want)
	}
}
