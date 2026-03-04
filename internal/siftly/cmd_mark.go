package siftly

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func (m *Model) markCurrent(colour ui.MarkColor) {
	if (m.cursor) < 0 || m.cursor >= len(m.table.filteredIndices) {
		return // This messed up as the cursor isn't at a point in the viewport
	}
	master := m.table.filteredIndices[m.cursor] // Gets the row
	id := m.table.rows[master].ID
	if colour == ui.MarkNone {
		delete(m.table.markedRows, id)
		logging.Infof("Cursor: %d with Stable ID %d has been unmarked", m.cursor, id)
	} else {
		logging.Infof("Cursor: %d with Stable ID %d is being marked with color %s", m.cursor, id, colour)
		m.table.markedRows[id] = colour
	}
}

func (m *Model) handleMarkCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.view.mode = modeView
		return m, nil

	case "r", "g", "a", "c":
		var mark ui.MarkColor
		switch msg.String() {
		case "r":
			mark = ui.MarkRed
		case "g":
			mark = ui.MarkGreen
		case "a":
			mark = ui.MarkAmber
		case "c":
			mark = ui.MarkNone
		}

		m.markCurrent(mark)
		m.view.mode = modeView

		m.refreshView("mark", false)

		// Notice only on actual change
		return m, m.view.notice.Start(
			fmt.Sprintf("Row %d marked [%s]", m.cursor+1, msg.String()),
			"",
			noticeDuration,
		)
	}

	// Unhandled keys: stay in mark mode, do nothing
	return m, nil
}

func (m *Model) jumpToNextMark() {
	logging.Debug("jumpToNextMark callled..")
	if !m.checkViewPortHasData() {
		return
	}

	for i := m.cursor + 1; i < len(m.table.filteredIndices); i++ {
		rowIdx := m.table.filteredIndices[i]
		row := m.table.rows[rowIdx]
		if _, ok := m.table.markedRows[row.ID]; ok {
			logging.Debugf("Next mark found at %d", i)
			m.cursor = i
			return
		}

	}
	logging.Debug("No next mark has been found")
}

func (m *Model) jumpToPreviousMark() {
	logging.Debug("jumpToPreviousMark called..")
	n := len(m.table.filteredIndices)
	if n == 0 {
		logging.Debug("filteredIndicies is emtpy")
	}
	if m.cursor < 0 {
		logging.Debug("Cursor at 0 or below")
	}

	for i := m.cursor - 1; i >= 0; i-- {
		rowIdx := m.table.filteredIndices[i]
		row := m.table.rows[rowIdx]
		if _, ok := m.table.markedRows[row.ID]; ok {
			logging.Debug("Previous mark has been found")
			m.cursor = i
			return
		}

	}
	logging.Debug("No previous mark has been found")
}
