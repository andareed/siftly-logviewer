package siftly

type rowRangeSelection struct {
	active bool
	anchor int
}

func (m *Model) toggleRowRangeSelection() bool {
	if m.view.rowRange.active {
		m.clearRowRangeSelection()
		return false
	}
	if m.cursor < 0 || m.cursor >= len(m.table.filteredIndices) {
		return false
	}
	m.view.rowRange = rowRangeSelection{active: true, anchor: m.cursor}
	return true
}

func (m *Model) clearRowRangeSelection() {
	m.view.rowRange = rowRangeSelection{}
}

func (m *Model) selectedDisplayRange() (start, end int, ok bool) {
	if !m.view.rowRange.active || len(m.table.filteredIndices) == 0 {
		return 0, 0, false
	}
	anchor := m.view.rowRange.anchor
	if anchor < 0 || anchor >= len(m.table.filteredIndices) || m.cursor < 0 || m.cursor >= len(m.table.filteredIndices) {
		return 0, 0, false
	}
	if anchor <= m.cursor {
		return anchor, m.cursor, true
	}
	return m.cursor, anchor, true
}

func (m *Model) actionDisplayRange() (start, end int, ok bool) {
	if start, end, ok := m.selectedDisplayRange(); ok {
		return start, end, true
	}
	if m.cursor < 0 || m.cursor >= len(m.table.filteredIndices) {
		return 0, 0, false
	}
	return m.cursor, m.cursor, true
}

func (m *Model) selectedRowCount() int {
	start, end, ok := m.selectedDisplayRange()
	if !ok {
		return 0
	}
	return end - start + 1
}

func (m *Model) rowIsRangeSelected(displayIndex int) bool {
	start, end, ok := m.selectedDisplayRange()
	return ok && displayIndex >= start && displayIndex <= end
}
