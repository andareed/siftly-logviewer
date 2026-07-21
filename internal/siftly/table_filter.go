package siftly

import (
	"regexp"
	"time"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

type filterJob struct {
	rows            []Row
	rowOrder        []int
	markedRows      map[uint64]ui.MarkColor
	showOnlyMarked  bool
	filterEnabled   bool
	filterRegex     *regexp.Regexp
	filterWholeRow  bool
	searchColumns   []int
	timeWindow      featuretimewindow.Window
	rowTimes        []time.Time
	rowHasTimes     []bool
	timeWindowOn    bool
	currentRowHash  uint64
	completionLabel string
}

func (m *Model) applyFilter() {
	logging.Debugf("applyFilter called")
	defer m.bumpGraphDataVersion()
	m.invalidateSearchIndex()
	m.clearRowRangeSelection()
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

func (j filterJob) run() []int {
	if !j.filterActive() && !j.showOnlyMarked && !j.timeWindowOn {
		out := make([]int, len(j.rowOrder))
		copy(out, j.rowOrder)
		return out
	}

	filtered := make([]int, 0, len(j.rowOrder))
	for _, rowIdx := range j.rowOrder {
		if rowIdx < 0 || rowIdx >= len(j.rows) {
			continue
		}
		row := j.rows[rowIdx]
		if j.includeRow(row, rowIdx) {
			filtered = append(filtered, rowIdx)
		}
	}
	return filtered
}

func (j filterJob) filterActive() bool {
	return j.filterEnabled && j.filterRegex != nil
}

func (j filterJob) includeRow(row Row, rowIndex int) bool {
	if j.showOnlyMarked {
		if _, ok := j.markedRows[row.ID]; !ok {
			return false
		}
	}

	if j.timeWindowOn {
		if rowIndex < 0 || rowIndex >= len(j.rowHasTimes) {
			return false
		}
		if !j.rowHasTimes[rowIndex] {
			return false
		}
		ts := j.rowTimes[rowIndex]
		if ts.Before(j.timeWindow.Start) || ts.After(j.timeWindow.End) {
			return false
		}
	}

	if j.filterActive() {
		match := row.MatchesColumns(j.filterRegex, j.searchColumns)
		if !match && j.filterWholeRow {
			match = j.filterRegex.MatchString(row.String())
		}
		if !match {
			return false
		}
	}
	return true
}
