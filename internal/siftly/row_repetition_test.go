package siftly

import "testing"

func TestRepeatedColumnsFollowDisplayedRowOrder(t *testing.T) {
	m := Model{
		table: tableState{
			rows: []Row{
				{Cols: []string{"host-a", "same", ""}},
				{Cols: []string{"host-b", "same", ""}},
				{Cols: []string{"host-b", "changed", "value"}},
			},
			filteredIndices: []int{1, 0, 2},
		},
	}

	if got := m.repeatedColumnsForDisplayedRow(0); got != nil {
		t.Fatalf("first displayed row repeats=%v want nil", got)
	}
	second := m.repeatedColumnsForDisplayedRow(1)
	if second[0] || !second[1] || second[2] {
		t.Fatalf("second displayed row repeats=%v", second)
	}
	third := m.repeatedColumnsForDisplayedRow(2)
	if third[0] || third[1] || third[2] {
		t.Fatalf("third displayed row repeats=%v", third)
	}
}
