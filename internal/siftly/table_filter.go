package siftly

import "github.com/andareed/siftly-hostlog/internal/shared/logging"

func (m *Model) applyFilter() {
	logging.Debugf("applyFilter called")
	defer m.bumpGraphDataVersion()
	// Remember the hash of what we have currently selected.
	currentRowHash := m.currentRowHashID()                // should be called before we reset the filtered indices
	m.table.filteredIndices = m.table.filteredIndices[:0] // reset slice

	filterActive := m.table.filterEnabled && m.table.filterRegex != nil
	if !filterActive && !m.table.showOnlyMarked && !m.table.timeWindow.Enabled {
		logging.Debug("applyFilter: no active filters; adding all indices")
		for i := range m.table.rows {
			m.table.filteredIndices = append(m.table.filteredIndices, i)
		}
		m.applyActiveSort()
		if len(m.table.filteredIndices) == 0 {
			m.cursor = 0
		}
		m.jumpToHashID(currentRowHash)
		m.clampCursor()
		return
	}

	for i, row := range m.table.rows {
		if m.includeRow(row, i) {
			m.table.filteredIndices = append(m.table.filteredIndices, i)
		}
	}
	m.applyActiveSort()

	if len(m.table.filteredIndices) == 0 {
		// No matches found; prevent index panics.
		m.cursor = -1
	}

	m.jumpToHashID(currentRowHash)
	m.clampCursor()
}
