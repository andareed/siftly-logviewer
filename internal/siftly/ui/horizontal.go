package ui

import (
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

const (
	frozenColumnSeparator = "│"
	tableGutterSeparator  = "│"
)

func composeHorizontalSegments(frozen, scrolling string, width, offset int) string {
	if width <= 0 {
		return ""
	}
	if offset < 0 {
		offset = 0
	}

	frozenLines := splitRenderLines(frozen)
	scrollingLines := splitRenderLines(scrolling)
	frozenWidth := renderLinesWidth(frozenLines)
	hasFrozen := frozenWidth > 0
	hasScrolling := renderLinesWidth(scrollingLines) > 0

	separatorWidth := 0
	if hasFrozen && hasScrolling && frozenWidth < width {
		separatorWidth = 1
	}
	if frozenWidth > width {
		frozenWidth = width
	}
	scrollingWidth := width - frozenWidth - separatorWidth
	if scrollingWidth < 0 {
		scrollingWidth = 0
	}

	height := max(len(frozenLines), len(scrollingLines))
	if height == 0 {
		height = 1
	}
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		fixed := ""
		if i < len(frozenLines) {
			fixed = cutAndPad(frozenLines[i], 0, frozenWidth)
		} else {
			fixed = strings.Repeat(" ", frozenWidth)
		}

		scrolled := ""
		if scrollingWidth > 0 && i < len(scrollingLines) {
			scrolled = cutAndPad(scrollingLines[i], offset, scrollingWidth)
		} else if scrollingWidth > 0 {
			scrolled = strings.Repeat(" ", scrollingWidth)
		}

		separator := ""
		if separatorWidth > 0 {
			separator = frozenColumnSeparator
		}
		lines[i] = fixed + separator + scrolled
	}
	return strings.Join(lines, "\n")
}

func splitRenderLines(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, "\n")
}

func renderLinesWidth(lines []string) int {
	width := 0
	for _, line := range lines {
		if lineWidth := xansi.StringWidth(line); lineWidth > width {
			width = lineWidth
		}
	}
	return width
}

func cutAndPad(value string, offset, width int) string {
	if width <= 0 {
		return ""
	}
	cut := xansi.Cut(value, offset, offset+width)
	missing := width - xansi.StringWidth(cut)
	if missing > 0 {
		cut += strings.Repeat(" ", missing)
	}
	return cut
}
