package siftly

import (
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestParseColumnTokens(t *testing.T) {
	tests := []struct {
		raw     string
		want    []string
		wantErr bool
	}{
		{raw: "1,2,3", want: []string{"1", "2", "3"}},
		{raw: "time details", want: []string{"time", "details"}},
		{raw: `"MAC Address",details`, want: []string{"MAC Address", "details"}},
		{raw: "'user name' notes", want: []string{"user name", "notes"}},
		{raw: `"unterminated`, wantErr: true},
	}

	for _, tt := range tests {
		got, err := parseColumnTokens(tt.raw)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("parseColumnTokens(%q) expected error, got nil", tt.raw)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseColumnTokens(%q) unexpected error: %v", tt.raw, err)
		}
		if len(got) != len(tt.want) {
			t.Fatalf("parseColumnTokens(%q) len=%d, want %d", tt.raw, len(got), len(tt.want))
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("parseColumnTokens(%q) token[%d]=%q want %q", tt.raw, i, got[i], tt.want[i])
			}
		}
	}
}

func TestResolveColumnIndex(t *testing.T) {
	m := &Model{
		table: tableState{
			header: []ui.ColumnMeta{
				{Name: "Time", Index: 0, Visible: true},
				{Name: "MAC Address", Index: 1, Visible: true},
				{Name: "Details", Index: 2, Visible: true},
			},
		},
	}

	tests := []struct {
		token string
		want  int
		ok    bool
	}{
		{token: "1", want: 0, ok: true},
		{token: "3", want: 2, ok: true},
		{token: "MAC Address", want: 1, ok: true},
		{token: "mac address", want: 1, ok: true},
		{token: "missing", ok: false},
	}

	for _, tt := range tests {
		got, ok := m.resolveColumnIndex(tt.token)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Fatalf("resolveColumnIndex(%q) = (%d,%t), want (%d,%t)", tt.token, got, ok, tt.want, tt.ok)
		}
	}
}

func TestToggleColumnsBySpec(t *testing.T) {
	m := &Model{
		table: tableState{
			header: []ui.ColumnMeta{
				{Name: "Time", Index: 0, Visible: true},
				{Name: "MAC Address", Index: 1, Visible: true},
				{Name: "Details", Index: 2, Visible: true},
			},
		},
	}

	toggled, missing, err := m.toggleColumnsBySpec("1,\"MAC Address\",missing")
	if err != nil {
		t.Fatalf("toggleColumnsBySpec returned error: %v", err)
	}
	if len(missing) != 1 || missing[0] != "missing" {
		t.Fatalf("missing=%v want [missing]", missing)
	}
	if len(toggled) != 2 {
		t.Fatalf("toggled len=%d want 2", len(toggled))
	}
	if m.table.header[0].Visible != false || m.table.header[1].Visible != false {
		t.Fatalf("columns not toggled off as expected: %+v", m.table.header)
	}
}

func TestColumnsNoticeText(t *testing.T) {
	if got := columnsNoticeText(nil, nil); got != "No columns matched" {
		t.Fatalf("empty -> %q", got)
	}
	if got := columnsNoticeText([]string{"A"}, nil); got != "Toggled: A" {
		t.Fatalf("toggled only -> %q", got)
	}
	if got := columnsNoticeText(nil, []string{"X"}); got != "Unknown: X" {
		t.Fatalf("missing only -> %q", got)
	}
	if got := columnsNoticeText([]string{"A"}, []string{"X"}); got != "Toggled: A · Unknown: X" {
		t.Fatalf("both -> %q", got)
	}
}
