package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type TimeWindowDrawerInput struct {
	Width         int
	HasTimeBounds bool
	StartInput    string
	EndInput      string
	ScrubberLine  string
	StepLabel     string
	ErrorMsg      string
	AreaStyle     lipgloss.Style
}

type TimeWindowScrubberInput struct {
	Width         int
	HasTimeBounds bool
	DraftStart    time.Time
	DraftEnd      time.Time
	TimeMin       time.Time
	TimeMax       time.Time
	TimeLayout    string
}

func FormatStep(step time.Duration) string {
	if step%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(step/time.Hour))
	}
	return fmt.Sprintf("%dm", int(step/time.Minute))
}

func RenderTimeWindowDrawer(in TimeWindowDrawerInput) string {
	innerWidth := max(0, in.Width-2)
	lineStyle := lipgloss.NewStyle().Width(innerWidth)

	startLine := "Start  " + in.StartInput
	endLine := "End    " + in.EndInput
	statusKind := "success"
	statusMsg := "✓ Draft window ready"
	switch {
	case in.ErrorMsg != "":
		statusKind = "error"
		statusMsg = "✖ " + in.ErrorMsg
	case !in.HasTimeBounds:
		statusKind = "warn"
		statusMsg = "⚠ No timestamps available"
	}

	statusStyle := lipgloss.NewStyle().Width(innerWidth)
	switch statusKind {
	case "success":
		statusStyle = statusStyle.Foreground(lipgloss.Color("42"))
	case "warn":
		statusStyle = statusStyle.Foreground(lipgloss.Color("214"))
	case "error":
		statusStyle = statusStyle.Foreground(lipgloss.Color("196"))
	}

	action := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render("[ Enter Apply ]")
	cancel := lipgloss.NewStyle().Faint(true).Render("Esc Cancel")
	reset := lipgloss.NewStyle().Faint(true).Render("R Reset")
	actionLine := action + "   " + cancel + "   " + reset
	actionPad := innerWidth - lipgloss.Width(actionLine)
	if actionPad > 0 {
		actionLine = strings.Repeat(" ", actionPad) + actionLine
	}
	controlsLine := lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf("Tab: next field  ←/→: move %s  Shift+←/→: expand %s  -/+: step", in.StepLabel, in.StepLabel),
	)

	lines := []string{
		lineStyle.Render("Window"),
		lineStyle.Render(in.ScrubberLine),
		lineStyle.Render(""),
		lineStyle.Render("Range"),
		lineStyle.Render(startLine),
		lineStyle.Render(endLine),
		lineStyle.Render(""),
		statusStyle.Render(statusMsg),
		lineStyle.Render(actionLine),
		lineStyle.Render(controlsLine),
	}

	content := strings.Join(lines, "\n")
	return in.AreaStyle.Width(in.Width).Render(content)
}

func RenderTimeWindowScrubber(in TimeWindowScrubberInput) string {
	if !in.HasTimeBounds {
		return "Scrubber: n/a"
	}

	start := in.DraftStart
	end := in.DraftEnd
	if start.IsZero() || end.IsZero() {
		start, end = in.TimeMin, in.TimeMax
	}

	minLabel := in.TimeMin.Format(in.TimeLayout)
	maxLabel := in.TimeMax.Format(in.TimeLayout)
	padding := 2
	barWidth := in.Width - len(minLabel) - len(maxLabel) - padding*2
	if barWidth < 10 {
		return fmt.Sprintf("Window: %s - %s", start.Format(in.TimeLayout), end.Format(in.TimeLayout))
	}

	bar := make([]rune, barWidth)
	for i := range bar {
		bar[i] = '-'
	}
	rangeDur := in.TimeMax.Sub(in.TimeMin)
	if rangeDur <= 0 {
		return "Scrubber: n/a"
	}

	windowStart := clampTimeToBounds(start, in.TimeMin, in.TimeMax)
	windowEnd := clampTimeToBounds(end, in.TimeMin, in.TimeMax)
	startPos := int(float64(barWidth-1) * windowStart.Sub(in.TimeMin).Seconds() / rangeDur.Seconds())
	endPos := int(float64(barWidth-1) * windowEnd.Sub(in.TimeMin).Seconds() / rangeDur.Seconds())
	if startPos < 0 {
		startPos = 0
	}
	if endPos >= barWidth {
		endPos = barWidth - 1
	}
	if endPos < startPos {
		startPos, endPos = endPos, startPos
	}
	for i := startPos; i <= endPos; i++ {
		bar[i] = '='
	}
	if startPos >= 0 && startPos < barWidth {
		bar[startPos] = '['
	}
	if endPos >= 0 && endPos < barWidth {
		bar[endPos] = ']'
	}

	return fmt.Sprintf("%s  %s  %s", minLabel, string(bar), maxLabel)
}

func RenderTimeWindowStatusLabel(enabled bool, start time.Time, end time.Time, layout string) string {
	if !enabled {
		return "Window: off"
	}
	return fmt.Sprintf("Window: %s - %s", start.Format(layout), end.Format(layout))
}

func clampTimeToBounds(t time.Time, min time.Time, max time.Time) time.Time {
	if t.Before(min) {
		return min
	}
	if t.After(max) {
		return max
	}
	return t
}
