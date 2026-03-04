package siftly

import (
	"time"

	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
)

type timeWindowResetBehavior int

const (
	timeWindowResetToDefault timeWindowResetBehavior = iota
	timeWindowResetDisable
)

const timeWindowResetMode = timeWindowResetDisable

func (m *Model) initTimeWindowState() {
	m.view.timeWindow = featuretimewindow.UIState{
		StartInput: featuretimewindow.InitInput(featuretimewindow.InputLayout),
		EndInput:   featuretimewindow.InitInput(featuretimewindow.InputLayout),
		Focus:      featuretimewindow.FocusStart,
	}

	m.computeTimeBounds()
	if m.table.timeWindow.Enabled && m.table.hasTimeBounds {
		m.table.timeWindow.Start = featuretimewindow.Clamp(m.table.timeWindow.Start, m.table.timeMin, m.table.timeMax)
		m.table.timeWindow.End = featuretimewindow.Clamp(m.table.timeWindow.End, m.table.timeMin, m.table.timeMax)
		if m.table.timeWindow.Start.After(m.table.timeWindow.End) {
			m.table.timeWindow.Start, m.table.timeWindow.End = featuretimewindow.DefaultBounds(m.table.timeMin, m.table.timeMax)
		}
	}
}

func (m *Model) computeTimeBounds() {
	header := make([]string, len(m.table.header))
	for i := range m.table.header {
		header[i] = m.table.header[i].Name
	}
	rows := make([][]string, len(m.table.rows))
	for i := range m.table.rows {
		rows[i] = m.table.rows[i].Cols
	}
	bounds := featuretimewindow.ComputeBounds(header, rows)

	m.table.timeColumnIndex = bounds.TimeColumnIndex
	m.table.rowTimes = bounds.RowTimes
	m.table.rowHasTimes = bounds.RowHasTimes
	m.table.hasTimeBounds = bounds.Has
	m.table.timeMin = bounds.Min
	m.table.timeMax = bounds.Max
}

func (m *Model) cursorTimestamp() (time.Time, bool) {
	return featuretimewindow.CursorTimestamp(
		m.table.filteredIndices,
		m.cursor,
		m.table.rowHasTimes,
		m.table.rowTimes,
	)
}

func (m *Model) setTimeWindowEdge(ts time.Time, setStart bool) {
	m.table.timeWindow = featuretimewindow.SetEdge(
		m.table.timeWindow,
		ts,
		m.table.timeMin,
		m.table.timeMax,
		setStart,
	)
	m.view.timeWindow.DraftStart = m.table.timeWindow.Start
	m.view.timeWindow.DraftEnd = m.table.timeWindow.End
	m.updateTimeWindowInputsFromDraft()
	m.applyFilter()
}
