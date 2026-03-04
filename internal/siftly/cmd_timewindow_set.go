package siftly

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleTimeWindowSetCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, m.exitCommand(true)
	case "b", "e":
		setStart := msg.String() == "b"
		if !m.table.hasTimeBounds {
			notice := m.view.notice.Start("No timestamps available", "warn", noticeDuration)
			return m, tea.Batch(m.exitCommand(true), notice)
		}

		ts, ok := m.cursorTimestamp()
		if !ok {
			notice := m.view.notice.Start("No timestamp on current row", "warn", noticeDuration)
			return m, tea.Batch(m.exitCommand(true), notice)
		}

		m.setTimeWindowEdge(ts, setStart)
		label := "end"
		if setStart {
			label = "start"
		}
		notice := m.view.notice.Start(fmt.Sprintf("Window %s set", label), "", noticeDuration)
		return m, tea.Batch(m.exitCommand(true), notice)
	case "r":
		if !m.table.hasTimeBounds {
			notice := m.view.notice.Start("No timestamps available", "warn", noticeDuration)
			return m, tea.Batch(m.exitCommand(true), notice)
		}

		m.table.timeWindow.Enabled = true
		m.table.timeWindow.Start = m.table.timeMin
		m.table.timeWindow.End = m.table.timeMax
		m.view.timeWindow.DraftStart = m.table.timeWindow.Start
		m.view.timeWindow.DraftEnd = m.table.timeWindow.End
		m.updateTimeWindowInputsFromDraft()
		m.applyFilter()

		notice := m.view.notice.Start("Window reset", "", noticeDuration)
		return m, tea.Batch(m.exitCommand(true), notice)
	}

	return m, nil
}
