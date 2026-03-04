package siftly

import (
	"errors"
	"time"

	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	sharedui "github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) openTimeWindowDrawer() {
	tw := &m.view.timeWindow
	tw.Open = true
	tw.ErrorMsg = ""
	tw.OrigWindow = m.table.timeWindow
	tw.Step = featuretimewindow.StepDefault

	if !m.table.hasTimeBounds {
		tw.ErrorMsg = "No timestamps available"
		tw.StartInput.SetValue("")
		tw.EndInput.SetValue("")
		tw.DraftStart = time.Time{}
		tw.DraftEnd = time.Time{}
		m.setTimeWindowFocus(featuretimewindow.FocusStart)
		m.view.mode = modeTimeWindow
		m.setModeHint("tab: next  enter: apply  r: reset  esc: cancel  ←/→ shift window  shift+←/→ expand  -/+ step")
		m.refreshView("time-window-open", true)
		return
	}

	if m.table.timeWindow.Start.IsZero() || m.table.timeWindow.End.IsZero() {
		tw.DraftStart, tw.DraftEnd = featuretimewindow.DefaultBounds(m.table.timeMin, m.table.timeMax)
	} else {
		tw.DraftStart = m.table.timeWindow.Start
		tw.DraftEnd = m.table.timeWindow.End
	}

	m.updateTimeWindowInputsFromDraft()
	m.setTimeWindowFocus(featuretimewindow.FocusStart)
	m.view.mode = modeTimeWindow
	m.setModeHint("tab: next  enter: apply  r: reset  esc: cancel  ←/→ shift window  shift+←/→ expand  -/+ step")
	m.refreshView("time-window-open", true)
}

func (m *Model) closeTimeWindowDrawer() {
	m.clearModeHint()
	m.view.timeWindow.Open = false
	m.view.timeWindow.ErrorMsg = ""
	m.view.mode = modeView
	m.refreshView("time-window-close", true)
}

func (m *Model) handleTimeWindowKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tw := &m.view.timeWindow

	switch {
	case msg.Type == tea.KeyEsc:
		m.closeTimeWindowDrawer()
		return m, nil
	case msg.Type == tea.KeyEnter:
		m.applyTimeWindowFromInputs()
		return m, nil
	case msg.String() == "r":
		m.resetTimeWindowDraft()
		return m, nil
	case msg.Type == tea.KeyTab:
		m.setTimeWindowFocus((tw.Focus + 1) % 3)
		return m, nil
	case msg.Type == tea.KeyShiftTab:
		m.setTimeWindowFocus((tw.Focus + 2) % 3)
		return m, nil
	case tw.Focus == featuretimewindow.FocusScrubber && msg.Type == tea.KeyLeft:
		m.shiftTimeWindow(-m.timeWindowStep())
		return m, nil
	case tw.Focus == featuretimewindow.FocusScrubber && msg.Type == tea.KeyRight:
		m.shiftTimeWindow(m.timeWindowStep())
		return m, nil
	case tw.Focus == featuretimewindow.FocusScrubber && msg.Type == tea.KeyShiftLeft:
		m.expandTimeWindow(-m.timeWindowStep())
		return m, nil
	case tw.Focus == featuretimewindow.FocusScrubber && msg.Type == tea.KeyShiftRight:
		m.expandTimeWindow(m.timeWindowStep())
		return m, nil
	case tw.Focus == featuretimewindow.FocusScrubber && msg.String() == "-":
		m.adjustTimeWindowStep(false)
		return m, nil
	case tw.Focus == featuretimewindow.FocusScrubber && (msg.String() == "+" || msg.String() == "="):
		m.adjustTimeWindowStep(true)
		return m, nil
	}

	var cmd tea.Cmd
	if tw.Focus == featuretimewindow.FocusStart {
		tw.StartInput, cmd = tw.StartInput.Update(msg)
		return m, cmd
	}
	if tw.Focus == featuretimewindow.FocusEnd {
		tw.EndInput, cmd = tw.EndInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) setTimeWindowFocus(focus int) {
	tw := &m.view.timeWindow
	tw.Focus = focus
	switch focus {
	case featuretimewindow.FocusStart:
		tw.StartInput.Focus()
		tw.EndInput.Blur()
	case featuretimewindow.FocusEnd:
		tw.StartInput.Blur()
		tw.EndInput.Focus()
	default:
		tw.StartInput.Blur()
		tw.EndInput.Blur()
	}
}

func (m *Model) updateTimeWindowInputsFromDraft() {
	tw := &m.view.timeWindow
	if !tw.DraftStart.IsZero() {
		tw.StartInput.SetValue(tw.DraftStart.Format(featuretimewindow.InputLayout))
	}
	if !tw.DraftEnd.IsZero() {
		tw.EndInput.SetValue(tw.DraftEnd.Format(featuretimewindow.InputLayout))
	}
}

func (m *Model) syncDraftFromInputs() {
	featuretimewindow.SyncDraftFromInputs(&m.view.timeWindow, m.table.hasTimeBounds, m.table.timeMax.Location())
}

func (m *Model) resetTimeWindowDraft() {
	tw := &m.view.timeWindow
	tw.ErrorMsg = ""

	if !m.table.hasTimeBounds {
		tw.ErrorMsg = "No timestamps available"
		return
	}

	tw.DraftStart, tw.DraftEnd = featuretimewindow.DefaultBounds(m.table.timeMin, m.table.timeMax)
	m.updateTimeWindowInputsFromDraft()

	if timeWindowResetMode == timeWindowResetDisable {
		m.table.timeWindow.Enabled = false
		m.applyFilter()
	}
}

func (m *Model) applyTimeWindowFromInputs() {
	tw := &m.view.timeWindow
	tw.ErrorMsg = ""

	if !m.table.hasTimeBounds {
		tw.ErrorMsg = "No timestamps available"
		return
	}

	loc := m.table.timeMax.Location()
	nextWindow, err := featuretimewindow.ParseInputWindow(
		tw.StartInput.Value(),
		tw.EndInput.Value(),
		loc,
		m.table.timeMin,
		m.table.timeMax,
	)
	if err != nil {
		switch {
		case errors.Is(err, featuretimewindow.ErrInvalidStart):
			tw.ErrorMsg = "Invalid start time"
		case errors.Is(err, featuretimewindow.ErrInvalidEnd):
			tw.ErrorMsg = "Invalid end time"
		default:
			tw.ErrorMsg = "Start is after end"
		}
		return
	}

	m.table.timeWindow = nextWindow
	tw.DraftStart = nextWindow.Start
	tw.DraftEnd = nextWindow.End
	m.applyFilter()
	m.closeTimeWindowDrawer()
}

func (m *Model) shiftTimeWindow(delta time.Duration) {
	tw := &m.view.timeWindow
	tw.ErrorMsg = ""

	if !m.table.hasTimeBounds {
		tw.ErrorMsg = "No timestamps available"
		return
	}

	m.syncDraftFromInputs()
	if tw.DraftStart.IsZero() || tw.DraftEnd.IsZero() {
		tw.DraftStart, tw.DraftEnd = featuretimewindow.DefaultBounds(m.table.timeMin, m.table.timeMax)
	}
	tw.DraftStart, tw.DraftEnd = featuretimewindow.ShiftRange(
		tw.DraftStart,
		tw.DraftEnd,
		m.table.timeMin,
		m.table.timeMax,
		delta,
	)
	m.updateTimeWindowInputsFromDraft()
}

func (m *Model) timeWindowStep() time.Duration {
	return featuretimewindow.NormalizeStep(m.view.timeWindow.Step)
}

func (m *Model) adjustTimeWindowStep(increase bool) {
	m.view.timeWindow.Step = featuretimewindow.AdjustStep(m.timeWindowStep(), increase)
}

func (m *Model) expandTimeWindow(delta time.Duration) {
	tw := &m.view.timeWindow
	tw.ErrorMsg = ""

	if !m.table.hasTimeBounds {
		tw.ErrorMsg = "No timestamps available"
		return
	}

	m.syncDraftFromInputs()
	if tw.DraftStart.IsZero() || tw.DraftEnd.IsZero() {
		tw.DraftStart, tw.DraftEnd = featuretimewindow.DefaultBounds(m.table.timeMin, m.table.timeMax)
	}
	tw.DraftStart, tw.DraftEnd = featuretimewindow.ExpandRange(
		tw.DraftStart,
		tw.DraftEnd,
		m.table.timeMin,
		m.table.timeMax,
		delta,
	)
	m.updateTimeWindowInputsFromDraft()
}

func (m *Model) timeWindowDrawerView(width int) string {
	tw := &m.view.timeWindow
	return sharedui.RenderTimeWindowDrawer(sharedui.TimeWindowDrawerInput{
		Width:         width,
		HasTimeBounds: m.table.hasTimeBounds,
		StartInput:    tw.StartInput.View(),
		EndInput:      tw.EndInput.View(),
		ScrubberLine:  m.timeWindowScrubberLine(max(0, width-2)),
		StepLabel:     sharedui.FormatStep(m.timeWindowStep()),
		ErrorMsg:      tw.ErrorMsg,
		AreaStyle:     lipgloss.NewStyle(),
	})
}

func (m *Model) timeWindowScrubberLine(width int) string {
	return sharedui.RenderTimeWindowScrubber(sharedui.TimeWindowScrubberInput{
		Width:         width,
		HasTimeBounds: m.table.hasTimeBounds,
		DraftStart:    m.view.timeWindow.DraftStart,
		DraftEnd:      m.view.timeWindow.DraftEnd,
		TimeMin:       m.table.timeMin,
		TimeMax:       m.table.timeMax,
		TimeLayout:    featuretimewindow.InputLayout,
	})
}

func (m *Model) timeWindowStatusLabel() string {
	return sharedui.RenderTimeWindowStatusLabel(
		m.table.timeWindow.Enabled,
		m.table.timeWindow.Start,
		m.table.timeWindow.End,
		featuretimewindow.InputLayout,
	)
}
