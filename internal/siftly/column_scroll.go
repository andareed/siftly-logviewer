package siftly

import (
	"fmt"
	"unicode/utf8"

	xansi "github.com/charmbracelet/x/ansi"
)

func (m *Model) tableMarkerWidth() int {
	return len(fmt.Sprintf("%d", len(m.table.rows))) +
		utf8.RuneCountInString(m.styles.PillMarker) +
		utf8.RuneCountInString(m.styles.CommentMarker) +
		1 // table gutter separator
}

func (m *Model) tableContentWidth() int {
	width := m.viewport.Width - m.tableMarkerWidth()
	if width < 1 {
		return 1
	}
	return width
}

func (m *Model) scrollColumns(delta int) {
	m.view.columnScrollOffset += delta
	m.clampColumnScrollOffset()
}

func (m *Model) clampColumnScrollOffset() {
	if m.view.columnScrollOffset < 0 {
		m.view.columnScrollOffset = 0
	}
	maximum := m.maxColumnScrollOffset()
	if m.view.columnScrollOffset > maximum {
		m.view.columnScrollOffset = maximum
	}
}

func (m *Model) maxColumnScrollOffset() int {
	contentWidth := m.tableContentWidth()
	frozenWidth := 0
	scrollingWidth := 0
	hasFrozen := false
	hasScrolling := false
	for _, column := range m.table.header {
		if !column.Visible || column.Width <= 0 {
			continue
		}
		width := xansi.StringWidth(m.styles.Cell.Width(column.Width).Render(""))
		if column.Frozen {
			frozenWidth += width
			hasFrozen = true
		} else {
			scrollingWidth += width
			hasScrolling = true
		}
	}

	if frozenWidth > contentWidth {
		frozenWidth = contentWidth
	}
	available := contentWidth - frozenWidth
	if hasFrozen && hasScrolling && available > 0 {
		available--
	}
	maximum := scrollingWidth - available
	if maximum < 0 {
		return 0
	}
	return maximum
}
