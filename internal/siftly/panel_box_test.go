package siftly

import (
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
)

func assertPanelShape(t *testing.T, out string, wantWidth, wantHeight int) []string {
	t.Helper()
	lines := strings.Split(out, "\n")
	if len(lines) != wantHeight {
		t.Fatalf("panel height mismatch: got %d want %d\n%s", len(lines), wantHeight, out)
	}
	for i, line := range lines {
		if got := xansi.StringWidth(line); got != wantWidth {
			t.Fatalf("panel width mismatch on line %d: got %d want %d\n%q", i, got, wantWidth, line)
		}
	}
	return lines
}

func TestRenderBoxedPanelWideFilterActive(t *testing.T) {
	out := renderBoxedPanel(
		"hostlog.json",
		panelStatusSpec{
			CurrentRow: 0,
			TotalRows:  0,
			Filter:     "My filter",
			MarksOn:    false,
		},
		[]string{
			"Time        Host        Details",
			strings.Repeat("─", 76),
			"row sample",
		},
		80,
		7,
	)

	lines := assertPanelShape(t, out, 80, 7)
	if !strings.Contains(lines[0], "hostlog.json") {
		t.Fatalf("missing filename in top border: %q", lines[0])
	}
	if !strings.Contains(lines[0], "Rows 0/0  Filter: My filter") {
		t.Fatalf("missing right status in top border: %q", lines[0])
	}
}

func TestTopBorderTruncatesFilterBeforeFilename(t *testing.T) {
	top := renderPanelTopBorder(
		"hostlog.json",
		panelStatusSpec{
			CurrentRow: 6,
			TotalRows:  300,
			Filter:     "severity=error and details contains a very long substring",
			MarksOn:    false,
		},
		58,
	)

	if !strings.Contains(top, "hostlog.json") {
		t.Fatalf("filename should remain intact before filename truncation: %q", top)
	}
	if !strings.Contains(top, "Filter:") || !strings.Contains(top, "…") {
		t.Fatalf("filter should truncate with ellipsis first: %q", top)
	}
}

func TestTopBorderDropsFilterBeforeMarksWhenNarrow(t *testing.T) {
	top := renderPanelTopBorder(
		"very_long_file_name.log",
		panelStatusSpec{
			CurrentRow: 6,
			TotalRows:  300,
			Filter:     "severity=error and host=alpha and detail=really-long-value",
			MarksOn:    true,
		},
		34,
	)

	if strings.Contains(top, "Filter:") {
		t.Fatalf("filter should be dropped first on narrow width: %q", top)
	}
	if !strings.Contains(top, "Marks: on") {
		t.Fatalf("marks should remain after filter drop when possible: %q", top)
	}
	if !strings.Contains(top, ".log") {
		t.Fatalf("filename extension should be preserved: %q", top)
	}
}

func TestRenderBoxedPanelOmitsInactiveFields(t *testing.T) {
	out := renderBoxedPanel(
		"hostlog.json",
		panelStatusSpec{
			CurrentRow: 6,
			TotalRows:  300,
			Filter:     "None",
			MarksOn:    false,
		},
		[]string{
			"Time        Host        Details",
			strings.Repeat("─", 56),
		},
		60,
		6,
	)

	lines := assertPanelShape(t, out, 60, 6)
	top := lines[0]
	if !strings.Contains(top, "Rows 6/300") {
		t.Fatalf("rows should always be present: %q", top)
	}
	if strings.Contains(top, "Filter:") {
		t.Fatalf("filter should be omitted when none: %q", top)
	}
	if strings.Contains(top, "Marks:") {
		t.Fatalf("marks should be omitted when off: %q", top)
	}
}
