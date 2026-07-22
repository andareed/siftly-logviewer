package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

type RowStyles struct {
	Row                lipgloss.Style
	RowSelected        lipgloss.Style
	RepeatedCell       lipgloss.Style
	Cell               lipgloss.Style
	RedMarker          lipgloss.Style
	GreenMarker        lipgloss.Style
	AmberMarker        lipgloss.Style
	SearchHighlight    lipgloss.Style
	RowTextFGColor     lipgloss.Color
	RowSelectedFG      lipgloss.Color
	RowSelectedBG      lipgloss.Color
	RowRangeSelectedBG lipgloss.Color
	DefaultMarker      string
	PillMarker         string
	CommentMarker      string
}

type RowRenderInput struct {
	Cols           []string
	OriginalIndex  int
	Selected       bool
	RangeSelected  bool
	SearchQuery    string
	TotalRows      int
	CommentPresent bool
	Mark           MarkColor
	ColsMeta       []ColumnMeta
	RepeatedCols   []bool
	ContentWidth   int
	ScrollOffset   int
	Styles         RowStyles
}

func RenderRowCells(cols []string, colsMeta []ColumnMeta, style lipgloss.Style) (string, int) {
	rendered := renderColumnCells(cols, colsMeta, style, rowCellOptions{})
	return rendered, lipgloss.Height(rendered)
}

type rowCellOptions struct {
	repeatedCols []bool
	repeatedCell lipgloss.Style
	searchQuery  string
	searchStyle  lipgloss.Style
}

func renderColumnCells(cols []string, colsMeta []ColumnMeta, style lipgloss.Style, options rowCellOptions) string {
	var rendered []string
	for _, meta := range colsMeta {
		if !meta.Visible || meta.Width <= 0 {
			continue
		}
		text := ""
		if meta.Index >= 0 && meta.Index < len(cols) {
			text = singleLineCellText(cols[meta.Index])
		}
		if strings.TrimSpace(options.searchQuery) != "" {
			text = highlightMatches(text, options.searchQuery, options.searchStyle)
		}
		textWidth := meta.Width - style.GetHorizontalFrameSize()
		if textWidth < 1 {
			textWidth = 1
		}
		text = xansi.Truncate(text, textWidth, "…")
		if meta.Index >= 0 && meta.Index < len(options.repeatedCols) && options.repeatedCols[meta.Index] {
			text = options.repeatedCell.Render(text)
		}
		rendered = append(rendered, style.Width(meta.Width).MaxHeight(1).Render(text))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func renderScrollableRowCells(cols []string, colsMeta []ColumnMeta, style lipgloss.Style, options rowCellOptions, width, offset int) (string, int) {
	frozenMeta := make([]ColumnMeta, 0, len(colsMeta))
	scrollingMeta := make([]ColumnMeta, 0, len(colsMeta))
	for _, meta := range colsMeta {
		if !meta.Visible || meta.Width <= 0 {
			continue
		}
		if meta.Frozen {
			frozenMeta = append(frozenMeta, meta)
		} else {
			scrollingMeta = append(scrollingMeta, meta)
		}
	}
	frozen := renderColumnCells(cols, frozenMeta, style, options)
	scrolling := renderColumnCells(cols, scrollingMeta, style, options)
	joined := composeHorizontalSegments(frozen, scrolling, width, offset)
	return joined, lipgloss.Height(joined)
}

func RenderRow(in RowRenderInput) (string, int) {
	rowBgStyle, rowFG, rowBG := resolveRowVisualStyle(in.Styles, in.Selected, in.RangeSelected)
	rowPrefix := bgSeq(rowBG) + fgSeq(rowFG)
	rowSuffix := termenv.CSI + "0m"

	standardMarker := getRowMarker(in.Mark, in.Styles)
	markerWidth := len(fmt.Sprintf("%d", in.TotalRows)) + utf8.RuneCountInString(in.Styles.CommentMarker)

	firstLineMarker := standardMarker + rowBgStyle.Render(fmt.Sprintf("%*d", markerWidth, in.OriginalIndex)+tableGutterSeparator)
	additionalLineMarker := standardMarker + rowBgStyle.Render(strings.Repeat(" ", markerWidth)+tableGutterSeparator)
	if in.CommentPresent {
		firstLineMarker = standardMarker + rowBgStyle.Render(
			in.Styles.CommentMarker+fmt.Sprintf("%*d", markerWidth-utf8.RuneCountInString(in.Styles.CommentMarker), in.OriginalIndex)+tableGutterSeparator,
		)
	}

	options := rowCellOptions{
		searchQuery: in.SearchQuery,
		searchStyle: in.Styles.SearchHighlight,
	}
	if !in.Selected && !in.RangeSelected && strings.TrimSpace(in.SearchQuery) == "" {
		options.repeatedCols = in.RepeatedCols
		options.repeatedCell = in.Styles.RepeatedCell
	}

	content := renderColumnCells(in.Cols, in.ColsMeta, in.Styles.Cell, options)
	rowHeight := lipgloss.Height(content)
	if in.ContentWidth > 0 {
		content, rowHeight = renderScrollableRowCells(in.Cols, in.ColsMeta, in.Styles.Cell, options, in.ContentWidth, in.ScrollOffset)
	}
	lines := strings.Split(content, "\n")

	for i := range lines {
		left := additionalLineMarker
		line := lines[i]
		line = restoreRowStyleAfterReset(line, rowPrefix)
		right := rowPrefix + line + rowSuffix
		if i == 0 {
			left = firstLineMarker
		}
		lines[i] = left + right
	}

	return strings.Join(lines, "\n"), rowHeight
}

func singleLineCellText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.ReplaceAll(value, "\n", " ↵ ")
}

func resolveRowVisualStyle(styles RowStyles, cursor, rangeSelected bool) (lipgloss.Style, lipgloss.Color, lipgloss.Color) {
	if cursor {
		return styles.RowSelected, styles.RowSelectedFG, styles.RowSelectedBG
	}
	if rangeSelected {
		return styles.RowSelected.
			Background(styles.RowRangeSelectedBG).
			Foreground(styles.RowTextFGColor), styles.RowTextFGColor, styles.RowRangeSelectedBG
	}
	return styles.Row, styles.RowTextFGColor, lipgloss.Color("")
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
