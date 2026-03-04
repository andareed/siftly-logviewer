package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HeaderColumn struct {
	Name    string
	Width   int
	Visible bool
}

func RenderHeader(markerWidth int, columns []HeaderColumn, cellStyle lipgloss.Style, headerStyle lipgloss.Style) string {
	var cells []string
	for _, col := range columns {
		if !col.Visible || col.Width <= 0 {
			continue
		}
		cell := cellStyle.Width(col.Width).Render(col.Name)
		cells = append(cells, cell)
	}
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	return headerStyle.Render(strings.Repeat(" ", markerWidth) + headerRow)
}
