package siftly

import "github.com/andareed/siftly-hostlog/internal/shared/logging"

func (m *Model) applyFilter() {
	logging.Debugf("applyFilter called")
	defer m.bumpGraphDataVersion()
	m.ensureTableDerivedState()
	// Remember the hash of what we have currently selected.
	currentRowHash := m.currentRowHashID()                // should be called before we reset the filtered indices
	m.table.filteredIndices = m.table.filteredIndices[:0] // reset slice
	rowOrder := m.table.rowOrder

	filterActive := m.table.filterEnabled && m.table.filterRegex != nil
	if !filterActive && !m.table.showOnlyMarked && !m.table.timeWindow.Enabled {
		logging.Debug("applyFilter: no active filters; adding all indices")
		m.table.filteredIndices = copyIntSlice(m.table.filteredIndices, rowOrder)
		if len(m.table.filteredIndices) == 0 {
			m.cursor = 0
		}
		m.jumpToHashID(currentRowHash)
		m.clampCursor()
		return
	}

	for _, rowIdx := range rowOrder {
		row := m.table.rows[rowIdx]
		if m.includeRow(row, rowIdx) {
			m.table.filteredIndices = append(m.table.filteredIndices, rowIdx)
		}
	}

	if len(m.table.filteredIndices) == 0 {
		// No matches found; prevent index panics.
		m.cursor = -1
	}

	m.jumpToHashID(currentRowHash)
	m.clampCursor()
}

func copyIntSlice(dst, src []int) []int {
	if len(src) == 0 {
		return dst[:0]
	}
	if cap(dst) < len(src) {
		dst = make([]int, len(src))
	} else {
		dst = dst[:len(src)]
	}
	copy(dst, src)
	return dst
}
