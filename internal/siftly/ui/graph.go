package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func RenderGraphMessage(width int, height int, msg string) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
}

func RenderLegendLine(width int, labels []string, colors []lipgloss.Color) string {
	if width <= 0 {
		return ""
	}
	if len(labels) == 0 {
		return strings.Repeat(" ", width)
	}

	var styles []lipgloss.Style
	if len(colors) > 0 {
		styles = make([]lipgloss.Style, len(colors))
		for i, c := range colors {
			styles[i] = lipgloss.NewStyle().Foreground(c)
		}
	}

	var parts []string
	for i, label := range labels {
		bullet := "●"
		if i < len(styles) {
			bullet = styles[i].Render(bullet)
		}
		parts = append(parts, bullet+" "+label)
	}
	line := strings.Join(parts, "  ")
	return truncatePlainLegend(line, width)
}

func truncatePlainLegend(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	if width <= 3 {
		return ansi.Truncate(s, width, "")
	}
	return ansi.Truncate(s, width, "...")
}
