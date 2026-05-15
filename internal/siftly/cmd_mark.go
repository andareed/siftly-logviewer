package siftly

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func (m *Model) markCurrent(colour ui.MarkColor) {
	m.markDisplayedRows(colour, 0)
}

func (m *Model) markDisplayedRows(colour ui.MarkColor, nextCount int) int {
	if (m.cursor) < 0 || m.cursor >= len(m.table.filteredIndices) {
		return 0 // This messed up as the cursor isn't at a point in the viewport
	}
	if nextCount < 0 {
		nextCount = 0
	}
	endCursor := m.cursor + nextCount
	if endCursor >= len(m.table.filteredIndices) {
		endCursor = len(m.table.filteredIndices) - 1
	}

	changed := 0
	for displayIdx := m.cursor; displayIdx <= endCursor; displayIdx++ {
		master := m.table.filteredIndices[displayIdx] // Gets the row
		id := m.table.rows[master].ID
		if colour == ui.MarkNone {
			delete(m.table.markedRows, id)
			logging.Infof("Cursor: %d with Stable ID %d has been unmarked", displayIdx, id)
		} else {
			logging.Infof("Cursor: %d with Stable ID %d is being marked with color %s", displayIdx, id, colour)
			m.table.markedRows[id] = colour
		}
		changed++
	}
	return changed
}

func (m *Model) handleMarkCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes && runesAreDigits(msg.Runes) {
		m.setCommandValue(m.commandValue() + string(msg.Runes))
		return m, nil
	}

	switch msg.String() {
	case "esc":
		_ = m.exitCommand(true)
		return m, nil

	case "backspace", "ctrl+h":
		m.setCommandValue(dropLastRune(m.commandValue()))
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

		nextCount, err := strconv.Atoi(strings.TrimSpace(m.commandValue()))
		if err != nil && strings.TrimSpace(m.commandValue()) != "" {
			return m, m.view.notice.Start("Invalid mark count", "warn", noticeDuration)
		}
		changed := m.markDisplayedRows(mark, nextCount)
		_ = m.exitCommand(false)

		m.refreshView("mark", false)

		return m, markNoticeCmd(m, changed, mark, msg.String())
	}

	// Unhandled keys: stay in mark mode, do nothing
	return m, nil
}

func runesAreDigits(runes []rune) bool {
	if len(runes) == 0 {
		return false
	}
	for _, r := range runes {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func dropLastRune(s string) string {
	if s == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

func markNoticeCmd(m *Model, changed int, mark ui.MarkColor, key string) tea.Cmd {
	if changed <= 0 {
		return m.view.notice.Start("No row marked", "warn", noticeDuration)
	}
	action := "marked [" + key + "]"
	if mark == ui.MarkNone {
		action = "cleared"
	}
	if changed == 1 {
		return m.view.notice.Start(
			fmt.Sprintf("Row %d %s", m.cursor+1, action),
			"",
			noticeDuration,
		)
	}
	return m.view.notice.Start(
		fmt.Sprintf("%d rows %s", changed, action),
		"",
		noticeDuration,
	)
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
