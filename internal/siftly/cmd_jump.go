package siftly

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
)

func (m *Model) jumpToStart() {
	logging.Debug("jumpToStart called...")
	if !m.checkViewPortHasData() {
		return
	}
	m.cursor = 0
}

func (m *Model) jumpToEnd() {
	logging.Debug("jumpToEnd called...")
	if !m.checkViewPortHasData() {
		return
	}
	// As filteredIndices gets fully populated with all rows if there is no
	// filter. We are safe to say this is the last one, regardless.
	m.cursor = len(m.table.filteredIndices) - 1
}

func (m *Model) jumpToLine(lineNo int) tea.Cmd {
	logging.Debug("jumpToLineNo")
	if !m.checkViewPortHasData() {
		return nil
	}
	if lineNo <= 0 {
		return m.view.notice.Start(fmt.Sprintf("Line %d out of bounds", lineNo), "warn", noticeDuration)
	}
	target := lineNo - 1
	if target >= len(m.table.rows) {
		return m.view.notice.Start(fmt.Sprintf("Line %d out of bounds", lineNo), "warn", noticeDuration)
	}
	for i, idx := range m.table.filteredIndices {
		if idx == target {
			m.cursor = i
			return nil
		}
	}
	return m.view.notice.Start(fmt.Sprintf("Line %d not in current filter", lineNo), "warn", noticeDuration)
}
