package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type RowStyles struct {
	Row             lipgloss.Style
	RowSelected     lipgloss.Style
	Cell            lipgloss.Style
	RedMarker       lipgloss.Style
	GreenMarker     lipgloss.Style
	AmberMarker     lipgloss.Style
	SearchHighlight lipgloss.Style
	RowTextFGColor  lipgloss.Color
	RowSelectedFG   lipgloss.Color
	RowSelectedBG   lipgloss.Color
	DefaultMarker   string
	PillMarker      string
	CommentMarker   string
}

type RowRenderInput struct {
	Cols           []string
	OriginalIndex  int
	Selected       bool
	SearchQuery    string
	TotalRows      int
	CommentPresent bool
	Mark           MarkColor
	ColsMeta       []ColumnMeta
	Styles         RowStyles
}

func RenderRowCells(cols []string, colsMeta []ColumnMeta, style lipgloss.Style) (string, int) {
	var rendered []string
	for i, text := range cols {
		if i >= len(colsMeta) {
			break
		}
		meta := colsMeta[i]
		if !meta.Visible || meta.Width <= 0 {
			continue
		}
		rendered = append(rendered, style.Width(meta.Width).Render(text))
	}
	joined := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	return joined, lipgloss.Height(joined)
}

func RenderRow(in RowRenderInput) (string, int) {
	rowBgStyle := in.Styles.Row
	rowPrefix := bgSeq(lipgloss.Color("")) + fgSeq(in.Styles.RowTextFGColor)
	if in.Selected {
		rowBgStyle = in.Styles.RowSelected
		rowPrefix = bgSeq(in.Styles.RowSelectedBG) + fgSeq(in.Styles.RowSelectedFG)
	}
	rowSuffix := termenv.CSI + "0m"

	standardMarker := getRowMarker(in.Mark, in.Styles)
	markerWidth := len(fmt.Sprintf("%d", in.TotalRows)) + utf8.RuneCountInString(in.Styles.CommentMarker)

	firstLineMarker := standardMarker + rowBgStyle.Render(fmt.Sprintf("%*d", markerWidth, in.OriginalIndex))
	additionalLineMarker := standardMarker + rowBgStyle.Render(strings.Repeat(" ", markerWidth))
	if in.CommentPresent {
		firstLineMarker = standardMarker + rowBgStyle.Render(
			in.Styles.CommentMarker+fmt.Sprintf("%*d", markerWidth-utf8.RuneCountInString(in.Styles.CommentMarker), in.OriginalIndex),
		)
	}

	cols := in.Cols
	if strings.TrimSpace(in.SearchQuery) != "" {
		highlighted := make([]string, len(cols))
		for i, col := range cols {
			highlighted[i] = highlightMatches(col, in.SearchQuery, in.Styles.SearchHighlight)
		}
		cols = highlighted
	}

	content, rowHeight := RenderRowCells(cols, in.ColsMeta, in.Styles.Cell)
	lines := strings.Split(content, "\n")

	for i := range lines {
		left := additionalLineMarker
		line := lines[i]
		if strings.TrimSpace(in.SearchQuery) != "" {
			line = restoreRowStyleAfterReset(line, rowPrefix)
		}
		right := rowPrefix + line + rowSuffix
		if i == 0 {
			left = firstLineMarker
		}
		lines[i] = left + right
	}

	return strings.Join(lines, "\n"), rowHeight
}

func getRowMarker(mark MarkColor, styles RowStyles) string {
	switch mark {
	case MarkRed:
		return styles.RedMarker.Render(styles.PillMarker)
	case MarkGreen:
		return styles.GreenMarker.Render(styles.PillMarker)
	case MarkAmber:
		return styles.AmberMarker.Render(styles.PillMarker)
	default:
		return styles.DefaultMarker
	}
}

func highlightMatches(text string, query string, hl lipgloss.Style) string {
	q := strings.TrimSpace(query)
	if q == "" || text == "" {
		return text
	}
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(q)
	var b strings.Builder
	start := 0
	for {
		idx := strings.Index(lowerText[start:], lowerQuery)
		if idx == -1 {
			b.WriteString(text[start:])
			break
		}
		idx += start
		b.WriteString(text[start:idx])
		match := text[idx : idx+len(lowerQuery)]
		b.WriteString(hl.Render(match))
		start = idx + len(lowerQuery)
	}
	return b.String()
}

func restoreRowStyleAfterReset(s string, rowPrefix string) string {
	if rowPrefix == "" {
		return s
	}
	reset := termenv.CSI + "0m"
	if !strings.Contains(s, reset) {
		return s
	}
	return strings.ReplaceAll(s, reset, reset+rowPrefix)
}

func fgSeq(c lipgloss.Color) string {
	return colorSeq(c, false)
}

func bgSeq(c lipgloss.Color) string {
	return colorSeq(c, true)
}

func colorSeq(c lipgloss.Color, bg bool) string {
	value := string(c)
	if value == "" {
		if bg {
			return termenv.CSI + "49m"
		}
		return termenv.CSI + "39m"
	}
	profile := lipgloss.ColorProfile()
	tc := profile.Color(value)
	if tc == nil {
		return ""
	}
	return termenv.CSI + tc.Sequence(bg) + "m"
}
