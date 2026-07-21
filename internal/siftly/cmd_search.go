package siftly

import (
	"fmt"
	mathbits "math/bits"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const searchIndexChunkSize = 20_000

type searchIndexState struct {
	id       int
	query    string
	bits     []uint64
	next     int
	total    int
	complete bool
}

type searchIndexChunkMsg struct {
	ID int
}

func (m *Model) setSearchQuery(query string) tea.Cmd {
	m.view.searchQuery = strings.TrimSpace(query)
	return m.startSearchIndex()
}

func (m *Model) startSearchIndex() tea.Cmd {
	id := m.view.searchIndex.id + 1
	query := strings.ToLower(strings.TrimSpace(m.view.searchQuery))
	state := searchIndexState{id: id, query: query}
	if query != "" {
		state.bits = make([]uint64, (len(m.table.filteredIndices)+63)/64)
		state.complete = len(m.table.filteredIndices) == 0
	}
	m.view.searchIndex = state
	if query == "" || state.complete {
		return nil
	}
	return searchIndexChunkCmd(id, 0)
}

func (m *Model) ensureSearchIndex() tea.Cmd {
	query := strings.ToLower(strings.TrimSpace(m.view.searchQuery))
	state := m.view.searchIndex
	if query == "" || (state.query == query && (state.next > 0 || state.complete)) {
		return nil
	}
	return m.startSearchIndex()
}

func (m *Model) invalidateSearchIndex() {
	m.view.searchIndex = searchIndexState{id: m.view.searchIndex.id + 1}
}

func searchIndexChunkCmd(id int, delay time.Duration) tea.Cmd {
	if delay <= 0 {
		return func() tea.Msg { return searchIndexChunkMsg{ID: id} }
	}
	return tea.Tick(delay, func(time.Time) tea.Msg { return searchIndexChunkMsg{ID: id} })
}

func (m *Model) handleSearchIndexChunk(msg searchIndexChunkMsg) tea.Cmd {
	state := &m.view.searchIndex
	if msg.ID != state.id || state.complete || state.query == "" {
		return nil
	}
	end := state.next + searchIndexChunkSize
	if end > len(m.table.filteredIndices) {
		end = len(m.table.filteredIndices)
	}
	for displayIdx := state.next; displayIdx < end; displayIdx++ {
		rowIdx := m.table.filteredIndices[displayIdx]
		if rowIdx < 0 || rowIdx >= len(m.table.rows) {
			continue
		}
		if rowContainsSearch(m.table.rows[rowIdx], state.query) {
			state.bits[displayIdx/64] |= uint64(1) << uint(displayIdx%64)
			state.total++
		}
	}
	state.next = end
	if state.next >= len(m.table.filteredIndices) {
		state.complete = true
		return nil
	}
	return searchIndexChunkCmd(state.id, time.Millisecond)
}

func (m *Model) searchStatusLabel() string {
	query := strings.ToLower(strings.TrimSpace(m.view.searchQuery))
	if query == "" {
		return ""
	}
	state := m.view.searchIndex
	if state.query != query {
		return "Search active"
	}
	if !state.complete {
		totalRows := len(m.table.filteredIndices)
		percent := 0
		if totalRows > 0 {
			percent = state.next * 100 / totalRows
		}
		return fmt.Sprintf("Indexing matches %d%%", percent)
	}
	if state.total == 0 {
		return "No matches"
	}
	if !searchBitSet(state.bits, m.cursor) {
		return fmt.Sprintf("%d matches", state.total)
	}
	return fmt.Sprintf("Match %d/%d", searchBitRank(state.bits, m.cursor), state.total)
}

func searchBitSet(words []uint64, index int) bool {
	if index < 0 || index/64 >= len(words) {
		return false
	}
	return words[index/64]&(uint64(1)<<uint(index%64)) != 0
}

func searchBitRank(words []uint64, index int) int {
	if index < 0 || len(words) == 0 {
		return 0
	}
	wordIndex := index / 64
	if wordIndex >= len(words) {
		wordIndex = len(words) - 1
		index = wordIndex*64 + 63
	}
	rank := 0
	for i := 0; i < wordIndex; i++ {
		rank += mathbits.OnesCount64(words[i])
	}
	maskBits := uint(index%64 + 1)
	mask := ^uint64(0)
	if maskBits < 64 {
		mask = (uint64(1) << maskBits) - 1
	}
	rank += mathbits.OnesCount64(words[wordIndex] & mask)
	return rank
}

func rowContainsSearch(row Row, lowerQuery string) bool {
	return strings.Contains(strings.ToLower(row.String()), lowerQuery)
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

	indexed := m.view.searchIndex.complete &&
		m.view.searchIndex.query == strings.ToLower(q) &&
		len(m.view.searchIndex.bits) == (n+63)/64
	for i := 0; i < n; i++ {
		idx := start + i*dir
		if idx < 0 {
			idx += n
		}
		if idx >= n {
			idx -= n
		}
		matched := indexed && searchBitSet(m.view.searchIndex.bits, idx)
		if !indexed {
			row := m.table.rows[m.table.filteredIndices[idx]]
			matched = rowContainsSearch(row, strings.ToLower(q))
		}
		if matched {
			m.cursor = idx
			return true
		}
	}
	return false
}
