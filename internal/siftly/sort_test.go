package siftly

import (
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestParseSortSpec(t *testing.T) {
	header := []ui.ColumnMeta{
		{Name: "Time", Index: 0, Visible: true},
		{Name: "MAC Address", Index: 1, Visible: true},
		{Name: "Hidden", Index: 2, Visible: false},
		{Name: "Details", Index: 3, Visible: true},
	}

	tests := []struct {
		name    string
		input   string
		want    sortSpec
		wantErr bool
	}{
		{
			name:  "name with spaces and desc",
			input: `MAC Address desc`,
			want:  sortSpec{Enabled: true, ColumnIndex: 1, Desc: true},
		},
		{
			name:  "quoted name with spaces and desc",
			input: `"MAC Address" desc`,
			want:  sortSpec{Enabled: true, ColumnIndex: 1, Desc: true},
		},
		{
			name:  "visible column number asc default",
			input: `3`,
			want:  sortSpec{Enabled: true, ColumnIndex: 3, Desc: false},
		},
		{
			name:  "visible column number desc",
			input: `2 desc`,
			want:  sortSpec{Enabled: true, ColumnIndex: 1, Desc: true},
		},
		{
			name:  "clear sort",
			input: `off`,
			want:  sortSpec{Enabled: false, ColumnIndex: 0, Desc: false},
		},
		{
			name:  "empty clears sort",
			input: ``,
			want:  sortSpec{Enabled: false, ColumnIndex: 0, Desc: false},
		},
		{
			name:  "row alias desc",
			input: `row desc`,
			want:  sortSpec{Enabled: true, ColumnIndex: sortColumnOriginalIndex, Desc: true},
		},
		{
			name:  "hash alias asc",
			input: `# asc`,
			want:  sortSpec{Enabled: true, ColumnIndex: sortColumnOriginalIndex, Desc: false},
		},
		{
			name:  "zero numeric alias asc",
			input: `0`,
			want:  sortSpec{Enabled: true, ColumnIndex: sortColumnOriginalIndex, Desc: false},
		},
		{
			name:    "unknown column",
			input:   `does-not-exist`,
			wantErr: true,
		},
		{
			name:    "unterminated quote",
			input:   `"bad desc`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSortSpec(tt.input, header)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseSortSpec(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSortSpec(%q) unexpected error: %v", tt.input, err)
			}
			if got.Enabled != tt.want.Enabled || got.ColumnIndex != tt.want.ColumnIndex || got.Desc != tt.want.Desc {
				t.Fatalf("parseSortSpec(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompareColumnValues(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  int // sign only
	}{
		{
			name:  "numeric compare",
			left:  "9",
			right: "11",
			want:  -1,
		},
		{
			name:  "timestamp compare",
			left:  "Mon Jul 21 08:16:55 BST 2025:5756256",
			right: "Mon Jul 21 08:17:55 BST 2025:5756256",
			want:  -1,
		},
		{
			name:  "string compare case insensitive",
			left:  "beta",
			right: "Gamma",
			want:  -1,
		},
		{
			name:  "equal values",
			left:  "same",
			right: "same",
			want:  0,
		},
	}

	sign := func(v int) int {
		switch {
		case v < 0:
			return -1
		case v > 0:
			return 1
		default:
			return 0
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sign(compareColumnValues(tt.left, tt.right))
			if got != tt.want {
				t.Fatalf("compareColumnValues(%q, %q) sign=%d, want %d", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestSetSortSpecKeepsFilteredRowsInSortedOrder(t *testing.T) {
	rows := []Row{
		{Cols: []string{"charlie", "ops"}, OriginalIndex: 1},
		{Cols: []string{"alpha", "ops"}, OriginalIndex: 2},
		{Cols: []string{"delta", "sales"}, OriginalIndex: 3},
	}
	for i := range rows {
		rows[i].ID = rows[i].ComputeID()
	}

	m := Model{
		table: tableState{
			header: []ui.ColumnMeta{
				{Name: "Name", Index: 0, Visible: true},
				{Name: "Team", Index: 1, Visible: true},
			},
			rows: rows,
		},
	}

	if err := m.setSortSpec("Name asc"); err != nil {
		t.Fatalf("set sort: %v", err)
	}
	if err := m.setFilterPattern(`ops`); err != nil {
		t.Fatalf("set filter: %v", err)
	}

	want := []int{1, 0}
	if len(m.table.filteredIndices) != len(want) {
		t.Fatalf("filtered row count = %d want %d", len(m.table.filteredIndices), len(want))
	}
	for i, idx := range want {
		if got := m.table.filteredIndices[i]; got != idx {
			t.Fatalf("filtered row %d = %d want %d", i, got, idx)
		}
	}
}
