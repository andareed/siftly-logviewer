package timewindow

import (
	"testing"
	"time"
)

func TestFindTimeColumnIndex(t *testing.T) {
	tests := []struct {
		name    string
		header  []string
		wantIdx int
	}{
		{
			name:    "hostlog time column",
			header:  []string{"Time", "Host", "Details"},
			wantIdx: 0,
		},
		{
			name:    "todaylog timestamp column",
			header:  []string{"date", "timestamp", "pid"},
			wantIdx: 1,
		},
		{
			name:    "fallback to date",
			header:  []string{"date", "pid", "value"},
			wantIdx: 0,
		},
		{
			name:    "no time column",
			header:  []string{"host", "details"},
			wantIdx: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindTimeColumnIndex(tt.header)
			if got != tt.wantIdx {
				t.Fatalf("FindTimeColumnIndex() = %d, want %d", got, tt.wantIdx)
			}
		})
	}
}

func TestParseLogTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{
			name:  "hostlog value with suffix",
			input: "Mon Jul 21 08:16:55 BST 2025:5756256",
			want:  time.Date(2025, time.July, 21, 8, 16, 55, 0, time.FixedZone("BST", 1*60*60)),
		},
		{
			name:  "hostlog fixed offset with suffix",
			input: "Wed Apr 08 13:20:43 GMT+08:00 2026:-2129437870",
			want:  time.Date(2026, time.April, 8, 13, 20, 43, 0, time.FixedZone("", 8*60*60)),
		},
		{
			name:  "hostlog CEST abbreviation",
			input: "Wed Apr 08 13:20:43 CEST 2026:-2129437870",
			want:  time.Date(2026, time.April, 8, 13, 20, 43, 0, time.FixedZone("", 2*60*60)),
		},
		{
			name:  "hostlog CAT abbreviation",
			input: "Wed Apr 08 13:20:43 CAT 2026:-2129437870",
			want:  time.Date(2026, time.April, 8, 13, 20, 43, 0, time.FixedZone("", 2*60*60)),
		},
		{
			name:  "hostlog arbitrary abbreviation",
			input: "Wed Apr 08 13:20:43 XYZ 2026:-2129437870",
			want:  time.Date(2026, time.April, 8, 13, 20, 43, 0, time.FixedZone("", 0)),
		},
		{
			name:  "unix seconds",
			input: "1769385627",
			want:  time.Unix(1769385627, 0),
		},
		{
			name:  "todaylog formatted date",
			input: "2026-01-26 06:00:27",
			want:  time.Date(2026, time.January, 26, 6, 0, 27, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseLogTimestamp(tt.input)
			if !ok {
				t.Fatalf("ParseLogTimestamp(%q) returned ok=false", tt.input)
			}
			if got.Unix() != tt.want.Unix() {
				t.Fatalf("ParseLogTimestamp(%q) unix=%d, want %d", tt.input, got.Unix(), tt.want.Unix())
			}
		})
	}
}

func TestParseZoneToken(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantName   string
		wantOffset int
		wantOK     bool
	}{
		{
			name:       "gmt fixed offset",
			input:      "GMT+08:00",
			wantName:   "GMT+08:00",
			wantOffset: 8 * 60 * 60,
			wantOK:     true,
		},
		{
			name:       "utc compact offset",
			input:      "UTC-0530",
			wantName:   "UTC-0530",
			wantOffset: -(5*60*60 + 30*60),
			wantOK:     true,
		},
		{
			name:       "known abbreviation",
			input:      "CEST",
			wantName:   "CEST",
			wantOffset: 2 * 60 * 60,
			wantOK:     true,
		},
		{
			name:       "arbitrary abbreviation",
			input:      "XYZ",
			wantName:   "XYZ",
			wantOffset: 0,
			wantOK:     true,
		},
		{
			name:   "invalid token",
			input:  "GMT+99:00",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotOffset, gotOK := parseZoneToken(tt.input)
			if gotOK != tt.wantOK {
				t.Fatalf("parseZoneToken(%q) ok=%t, want %t", tt.input, gotOK, tt.wantOK)
			}
			if !gotOK {
				return
			}
			if gotName != tt.wantName || gotOffset != tt.wantOffset {
				t.Fatalf("parseZoneToken(%q) = (%q, %d), want (%q, %d)", tt.input, gotName, gotOffset, tt.wantName, tt.wantOffset)
			}
		})
	}
}
