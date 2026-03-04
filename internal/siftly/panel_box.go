package siftly

import (
	"fmt"
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
)

const (
	footerRows        = 3
	panelChromeRows   = 4 // top border, header row, separator row, bottom border
	drawerChromeRows  = 2 // top border, bottom border
	drawerContentRows = 8
	panelMinOuterCols = 6 // allows "│ " + at least 2 cells + " │"
)

type panelStatusSpec struct {
	CurrentRow int
	TotalRows  int
	Filter     string
	MarksOn    bool
	RightText  string
}

// renderBoxedPanel renders a fixed-size boxed panel with a status-rich top border.
func renderBoxedPanel(titleLeft string, status panelStatusSpec, innerLines []string, width, height int) string {
	if width < panelMinOuterCols {
		width = panelMinOuterCols
	}
	if height < 2 {
		height = 2
	}

	top := renderPanelTopBorder(titleLeft, status, width)
	bottom := renderPanelBottomBorder(width)

	innerRows := height - 2
	innerWidth := width - 4
	if innerWidth < 1 {
		innerWidth = 1
	}

	lines := make([]string, 0, height)
	lines = append(lines, top)
	for i := 0; i < innerRows; i++ {
		content := ""
		if i < len(innerLines) {
			content = innerLines[i]
		}
		lines = append(lines, renderPanelInnerRow(content, innerWidth))
	}
	lines = append(lines, bottom)
	return strings.Join(lines, "\n")
}

func renderPanelTopBorder(titleLeft string, status panelStatusSpec, width int) string {
	if width < panelMinOuterCols {
		return "┌┐"
	}

	title := strings.TrimSpace(titleLeft)
	if title == "" {
		title = "(untitled)"
	}

	filterRaw := strings.TrimSpace(status.Filter)
	includeFilter := filterRaw != "" && !strings.EqualFold(filterRaw, "none")
	includeMarks := status.MarksOn

	filterLimit := len([]rune(filterRaw))
	titleLimit := len([]rune(title))
	if titleLimit < 1 {
		titleLimit = 1
	}
	fullTitleLimit := titleLimit

	innerBudget := width - 2

	for {
		filterValue := ""
		if includeFilter {
			filterValue = truncateEndRunes(filterRaw, filterLimit)
		}

		left := truncateFilenameMiddlePreserveExt(title, titleLimit)
		right := buildPanelRightStatusWithOverride(
			status.CurrentRow,
			status.TotalRows,
			filterValue,
			includeFilter,
			includeMarks,
			status.RightText,
		)

		leftW := xansi.StringWidth(left)
		rightW := xansi.StringWidth(right)
		fillerLen := innerBudget - (leftW + rightW + 4) // spaces around left/filler/right and edges
		if fillerLen >= 1 {
			return "┌ " + left + " " + strings.Repeat("─", fillerLen) + " " + right + " ┐"
		}

		switch {
		case includeFilter && filterLimit > 1:
			// 1) shrink filter value first
			filterLimit--
		case titleLimit > 1:
			// 2) then shrink filename (middle ellipsis, keep extension)
			titleLimit--
		case includeFilter:
			// 3) if still too narrow, drop Filter
			includeFilter = false
			titleLimit = fullTitleLimit
		case includeMarks:
			// 4) then drop Marks
			includeMarks = false
			titleLimit = fullTitleLimit
		default:
			// Last resort: plain border if terminal is extremely narrow.
			return "┌" + strings.Repeat("─", width-2) + "┐"
		}
	}
}

func buildPanelRightStatusWithOverride(current, total int, filter string, includeFilter, includeMarks bool, rightOverride string) string {
	if strings.TrimSpace(rightOverride) != "" {
		return rightOverride
	}
	parts := []string{fmt.Sprintf("Rows %d/%d", current, total)}
	if includeFilter {
		parts = append(parts, "Filter: "+filter)
	}
	if includeMarks {
		parts = append(parts, "Marks: on")
	}
	return strings.Join(parts, "  ")
}

func renderPanelInnerRow(content string, innerWidth int) string {
	if innerWidth < 1 {
		return "││"
	}
	clipped := xansi.Truncate(content, innerWidth, "")
	pad := innerWidth - xansi.StringWidth(clipped)
	if pad < 0 {
		pad = 0
	}
	return "│ " + clipped + strings.Repeat(" ", pad) + " │"
}

func renderPanelBottomBorder(width int) string {
	if width < 2 {
		return ""
	}
	return "└" + strings.Repeat("─", width-2) + "┘"
}

func splitContentLines(s string) []string {
	trimmed := strings.TrimSuffix(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}
