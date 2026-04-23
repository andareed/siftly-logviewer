package siftly

import (
	"regexp"
	"testing"
)

func mustCompileFilter(t *testing.T, pattern string) *regexp.Regexp {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("compile %q: %v", pattern, err)
	}
	return re
}

func testRows() []Row {
	rows := []Row{
		{Cols: []string{"severity=error", "host=a"}},
		{Cols: []string{"severity=info", "host=b"}},
	}
	for i := range rows {
		rows[i].ID = rows[i].ComputeID()
		rows[i].OriginalIndex = i + 1
	}
	return rows
}

func TestToggleFilterDisablesAndReEnablesConfiguredPattern(t *testing.T) {
	m := Model{
		table: tableState{
			rows: testRows(),
		},
	}

	if err := m.setFilterPattern(`severity=error`); err != nil {
		t.Fatalf("set filter: %v", err)
	}
	if !m.table.filterEnabled {
		t.Fatalf("filter should start enabled")
	}
	if got := len(m.table.filteredIndices); got != 1 {
		t.Fatalf("filtered rows after apply = %d want 1", got)
	}

	if !m.toggleFilter() {
		t.Fatalf("toggle should succeed for configured filter")
	}
	if m.table.filterEnabled {
		t.Fatalf("filter should be disabled after first toggle")
	}
	if got := len(m.table.filteredIndices); got != 2 {
		t.Fatalf("filtered rows after disable = %d want 2", got)
	}

	if !m.toggleFilter() {
		t.Fatalf("toggle should re-enable configured filter")
	}
	if !m.table.filterEnabled {
		t.Fatalf("filter should be enabled after second toggle")
	}
	if got := len(m.table.filteredIndices); got != 1 {
		t.Fatalf("filtered rows after re-enable = %d want 1", got)
	}
}

func TestToggleFilterFailsWithoutConfiguredPattern(t *testing.T) {
	m := Model{
		table: tableState{
			rows: testRows(),
		},
	}

	if m.toggleFilter() {
		t.Fatalf("toggle should fail when no filter is configured")
	}
}

func TestClearFilterResetsPatternAndDisabledState(t *testing.T) {
	m := Model{
		table: tableState{
			rows: testRows(),
		},
	}

	if err := m.setFilterPattern(`severity=error`); err != nil {
		t.Fatalf("set filter: %v", err)
	}
	m.clearFilter()

	if m.table.filterPattern != "" {
		t.Fatalf("filter pattern should be cleared, got %q", m.table.filterPattern)
	}
	if m.table.filterEnabled {
		t.Fatalf("filter should be disabled after clear")
	}
	if m.table.filterRegex != nil {
		t.Fatalf("compiled filter should be nil after clear")
	}
	if got := len(m.table.filteredIndices); got != 2 {
		t.Fatalf("filtered rows after clear = %d want 2", got)
	}
}
