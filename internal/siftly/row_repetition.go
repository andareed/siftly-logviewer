package siftly

import "strings"

func (m *Model) repeatedColumnsForDisplayedRow(filteredIndex int) []bool {
	if filteredIndex <= 0 || filteredIndex >= len(m.table.filteredIndices) {
		return nil
	}
	currentIndex := m.table.filteredIndices[filteredIndex]
	previousIndex := m.table.filteredIndices[filteredIndex-1]
	if currentIndex < 0 || currentIndex >= len(m.table.rows) ||
		previousIndex < 0 || previousIndex >= len(m.table.rows) {
		return nil
	}

	current := m.table.rows[currentIndex].Cols
	previous := m.table.rows[previousIndex].Cols
	repeated := make([]bool, len(current))
	for sourceIndex, value := range current {
		if sourceIndex >= len(previous) || strings.TrimSpace(value) == "" {
			continue
		}
		repeated[sourceIndex] = value == previous[sourceIndex]
	}
	return repeated
}
