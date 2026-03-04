package siftly

import (
	"regexp"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
)

func (m *Model) setFilterPattern(pattern string) error {
	logging.Infof("Setting Pattern to: %s", pattern)
	if pattern == "" {
		m.table.filterRegex = nil
	} else {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		m.table.filterRegex = re
	}
	m.applyFilter()
	return nil
}

// region Filtering

func (m *Model) includeRow(row Row, rowIndex int) bool {
	return m.includeRowWithFilter(row, rowIndex, m.table.filterRegex)
}

func (m *Model) includeRowWithFilter(row Row, rowIndex int, re *regexp.Regexp) bool {
	if m.table.showOnlyMarked {
		if _, ok := m.table.markedRows[row.ID]; !ok {
			return false
		}
	}

	if m.table.timeWindow.Enabled {
		if rowIndex < 0 || rowIndex >= len(m.table.rowHasTimes) {
			return false
		}
		if !m.table.rowHasTimes[rowIndex] {
			return false
		}
		ts := m.table.rowTimes[rowIndex]
		if ts.Before(m.table.timeWindow.Start) || ts.After(m.table.timeWindow.End) {
			return false
		}
	}

	if re != nil {
		match := re.MatchString(row.String())
		if !match {
			return false
		}
	}
	return true
}
