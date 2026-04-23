package siftly

import (
	"regexp"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
)

func (m *Model) setFilterPattern(pattern string) error {
	logging.Infof("Setting Pattern to: %s", pattern)
	if pattern == "" {
		m.clearFilter()
	} else {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		m.table.filterPattern = pattern
		m.table.filterEnabled = true
		m.table.filterRegex = re
		m.applyFilter()
	}
	return nil
}

func (m *Model) clearFilter() {
	m.table.filterPattern = ""
	m.table.filterEnabled = false
	m.table.filterRegex = nil
	m.applyFilter()
}

func (m *Model) toggleFilter() bool {
	if strings.TrimSpace(m.table.filterPattern) == "" || m.table.filterRegex == nil {
		return false
	}
	m.table.filterEnabled = !m.table.filterEnabled
	m.applyFilter()
	return true
}

func (m *Model) filterStatusValue() string {
	pattern := strings.TrimSpace(m.table.filterPattern)
	if pattern == "" {
		return ""
	}
	state := "off"
	if m.table.filterEnabled && m.table.filterRegex != nil {
		state = "on"
	}
	return pattern + " (" + state + ")"
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

	if m.table.filterEnabled && re != nil {
		match := re.MatchString(row.String())
		if !match {
			return false
		}
	}
	return true
}
