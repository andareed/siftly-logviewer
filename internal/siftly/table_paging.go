package siftly

func (m *Model) pageDown() {
	if len(m.table.filteredIndices) == 0 {
		return
	}
	if m.cursor+m.pageRowSize < len(m.table.filteredIndices) {
		m.cursor += m.pageRowSize
	} else {
		m.cursor = len(m.table.filteredIndices) - 1
	}
}

func (m *Model) pageUp() {
	if len(m.table.filteredIndices) == 0 {
		return
	}
	m.cursor -= m.pageRowSize
	if m.cursor < 0 {
		m.cursor = 0
	}
}
