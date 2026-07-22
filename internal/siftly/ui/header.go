package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

type HeaderColumn struct {
	Name    string
	Width   int
	Visible bool
	Frozen  bool
}

func RenderHeader(markerWidth int, columns []HeaderColumn, contentWidth int, scrollOffset int, cellStyle lipgloss.Style, headerStyle lipgloss.Style) string {
	frozen := renderHeaderCells(columns, true, cellStyle)
	scrolling := renderHeaderCells(columns, false, cellStyle)
	headerRow := composeHorizontalSegments(frozen, scrolling, contentWidth, scrollOffset)
	gutter := ""
	if markerWidth > 0 {
		gutter = strings.Repeat(" ", markerWidth-1) + tableGutterSeparator
	}
	return headerStyle.Render(gutter + headerRow)
}

func renderHeaderCells(columns []HeaderColumn, frozen bool, cellStyle lipgloss.Style) string {
	var cells []string
	for _, col := range columns {
		if !col.Visible || col.Width <= 0 || col.Frozen != frozen {
			continue
		}
		textWidth := col.Width - cellStyle.GetHorizontalFrameSize()
		if textWidth < 1 {
			textWidth = 1
		}
		name := xansi.Truncate(singleLineCellText(col.Name), textWidth, "…")
		cell := cellStyle.Width(col.Width).MaxHeight(1).Render(name)
		cells = append(cells, cell)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}
