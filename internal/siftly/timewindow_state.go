package siftly

import (
	"time"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
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
	if m.table.derivedTimeData && len(m.table.rowTimes) == len(m.table.rows) && len(m.table.rowHasTimes) == len(m.table.rows) {
		return
	}

	defer logging.TimeOperation("compute time bounds")()

	header := make([]string, len(m.table.header))
	for i := range m.table.header {
		header[i] = m.table.header[i].Name
	}
	timeCol := featuretimewindow.FindTimeColumnIndex(header)
	rowTimes := make([]time.Time, len(m.table.rows))
	rowHasTimes := make([]bool, len(m.table.rows))

	hasAny := false
	var minTime time.Time
	var maxTime time.Time
	if timeCol >= 0 {
		for i, row := range m.table.rows {
			if timeCol >= len(row.Cols) {
				continue
			}
			ts, ok := featuretimewindow.ParseLogTimestamp(row.Cols[timeCol])
			if !ok {
				continue
			}
			rowTimes[i] = ts
			rowHasTimes[i] = true
			if !hasAny {
				minTime = ts
				maxTime = ts
				hasAny = true
				continue
			}
			if ts.Before(minTime) {
				minTime = ts
			}
			if ts.After(maxTime) {
				maxTime = ts
			}
		}
	}

	m.table.timeColumnIndex = timeCol
	m.table.rowTimes = rowTimes
	m.table.rowHasTimes = rowHasTimes
	m.table.hasTimeBounds = hasAny
	m.table.timeMin = minTime
	m.table.timeMax = maxTime
	m.table.derivedTimeData = true
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
	previous := m.table.timeWindow
	m.table.timeWindow = featuretimewindow.SetEdge(
		m.table.timeWindow,
		ts,
		m.table.timeMin,
		m.table.timeMax,
		setStart,
	)
	if previous != m.table.timeWindow {
		m.markDirty()
	}
	m.view.timeWindow.DraftStart = m.table.timeWindow.Start
	m.view.timeWindow.DraftEnd = m.table.timeWindow.End
	m.updateTimeWindowInputsFromDraft()
	m.applyFilter()
}
