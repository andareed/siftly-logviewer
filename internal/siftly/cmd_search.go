package siftly

import "strings"

func (m *Model) setSearchQuery(query string) {
	m.view.searchQuery = strings.TrimSpace(query)
}

func (m *Model) searchNext() bool {
	return m.searchFrom(m.cursor+1, 1)
}

func (m *Model) searchPrev() bool {
	return m.searchFrom(m.cursor-1, -1)
}

func (m *Model) searchFrom(start int, dir int) bool {
	q := strings.TrimSpace(m.view.searchQuery)
	if q == "" || len(m.table.filteredIndices) == 0 {
		return false
	}

	n := len(m.table.filteredIndices)
	if start < 0 {
		start = n - 1
	}
	if start >= n {
		start = 0
	}

	for i := 0; i < n; i++ {
		idx := start + i*dir
		if idx < 0 {
			idx += n
		}
		if idx >= n {
			idx -= n
		}
		row := m.table.rows[m.table.filteredIndices[idx]]
		if strings.Contains(strings.ToLower(row.String()), strings.ToLower(q)) {
			m.cursor = idx
			return true
		}
	}
	return false
}
