package siftly

import "github.com/andareed/siftly-hostlog/internal/shared/logging"

func (m *Model) clampCursor() {
	if len(m.table.filteredIndices) == 0 {
		m.cursor = -1
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.table.filteredIndices) {
		m.cursor = len(m.table.filteredIndices) - 1
	}
}

func (m *Model) currentRowHashID() uint64 {
	if len(m.table.filteredIndices) == 0 || m.cursor < 0 || m.cursor >= len(m.table.filteredIndices) {
		logging.Debugf("currentRowHashID called but no filteredIndices available (cursor=%d, len=%d)", m.cursor, len(m.table.filteredIndices))
		return 0 // or some sentinel value
	}
	rowIdx := m.table.filteredIndices[m.cursor]
	hashID := m.table.rows[rowIdx].ID
	logging.Debugf("currentRowHashID called returning HashID[%d] for cursor[%d] at filteredIndex[%d] which maps to rowIndex[%d]", hashID, m.cursor, rowIdx, rowIdx)
	return hashID
}

func (m *Model) jumpToHashID(hashID uint64) {
	if hashID == 0 {
		logging.Debugf("jumpToHashID called with HashID of 0, so returning")
		m.cursor = 0
		return
	}

	logging.Debugf("jumpToHashID called looking for HashID[%d]", hashID)
	for i, idx := range m.table.filteredIndices {
		if m.table.rows[idx].ID == hashID {
			m.cursor = i
			logging.Debugf("jumpToHashID: Jumping to index [%d] for hashID[%d]", i, hashID)
			return
		}
	}

	m.cursor = 0
	logging.Warnf("jumpToHashID: No match found for hashID[%d] so setting cursor to 0", hashID)
}

func (m *Model) checkViewPortHasData() bool {
	if len(m.table.filteredIndices) == 0 {
		logging.Debug("filterIndicies is empty")
		return false
	}
	if m.cursor < 0 {
		logging.Debug("Cursor at 0 or below")
		return false
	}
	return true
}
