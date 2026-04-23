package siftly

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) headerView() string {
	return m.renderHeaderLine(m.styles.Header)
}

func (m *Model) panelHeaderView() string {
	return m.renderHeaderLine(lipgloss.NewStyle())
}

func (m *Model) renderHeaderLine(headerStyle lipgloss.Style) string {
	cols := make([]ui.HeaderColumn, 0, len(m.table.header))
	showIndices := m.view.mode == modeCommand &&
		(m.view.command.cmd == CmdSort || m.view.command.cmd == CmdColumns)

	visibleIdx := 0
	for _, col := range m.table.header {
		if !col.Visible || col.Width <= 0 {
			continue
		}
		visibleIdx++
		name := m.sortedHeaderName(col)
		if showIndices {
			name = fmt.Sprintf("%d:%s", visibleIdx, name)
		}
		cols = append(cols, ui.HeaderColumn{
			Name:    name,
			Width:   col.Width,
			Visible: col.Visible,
		})
	}

	markerWidth := len(fmt.Sprintf("%d", len(m.table.rows))) +
		utf8.RuneCountInString(m.styles.PillMarker) +
		utf8.RuneCountInString(m.styles.CommentMarker)
	return ui.RenderHeader(markerWidth, cols, m.styles.Cell, headerStyle)
}

func (m *Model) footerView(width int) string {
	logging.Debugf("footerView mode=%d cmd=%d", m.view.mode, m.view.command.cmd)

	footerMode := CmdNone
	modeInput := ""
	isInputMode := false
	switch m.view.mode {
	case modeView:
		footerMode = CmdNone
	case modeComment:
		footerMode = CmdComment
		isInputMode = true
	case modeCommand:
		isInputMode = true
		footerMode = m.view.command.cmd
		modeInput = m.view.command.buf
	case modeTimeWindow:
		footerMode = CmdTimeWindowSet
		isInputMode = true
	}

	hints := "v view · c comment · t time · / search · ? help"
	if m.graphConfig.Enabled {
		hints += " · w graph"
	}

	debugInfo := ""
	if logging.IsDebugMode() && !isInputMode {
		debugInfo = fmt.Sprintf("dbg term=%dx%d vp=%dx%d cur=%d vis=%d-%d page=%d ch=%d hf=%d abv=%d",
			m.terminalWidth, m.terminalHeight, m.viewport.Width, m.viewport.Height,
			m.cursor, m.view.visibleStart, m.view.visibleEnd, m.pageRowSize,
			m.view.debugCursorHeight, m.view.debugHeightFree, m.view.debugDesiredAboveHeight,
		)
	}

	status := ""
	if m.view.notice.Msg != "" {
		status = ui.NoticeText(m.view.notice.Msg, m.view.notice.Type)
	}
	if status == "" {
		status = m.timeWindowStatusLabel()
	}
	if !isInputMode && debugInfo != "" {
		if strings.TrimSpace(status) == "" {
			status = debugInfo
		} else {
			status = status + " | " + debugInfo
		}
	}

	modeBanner := commandLabel(footerMode) + " MODE"
	if m.view.mode == modeTimeWindow {
		modeBanner = "TIME WINDOW MODE"
	}
	return ui.RenderFooter(width, ui.FooterState{
		ModeLabel:     modeBanner,
		StatusMessage: status,
		Hints:         hints,
		IsInputMode:   isInputMode,
		Prompt:        modeInput,
	}, ui.DefaultFooterStyles())
}

func (m *Model) View() string {
	if !m.ready {
		return "loading..."
	}

	if m.activeDialog != nil && m.activeDialog.IsVisible() {
		w, h := m.terminalWidth, m.terminalHeight
		return lipgloss.Place(
			w, h,
			lipgloss.Center, lipgloss.Center,
			m.activeDialog.View(),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceBackground(lipgloss.Color("236")),
		)
	}

	panelW := m.viewport.Width + 4
	if panelW < panelMinOuterCols {
		panelW = panelMinOuterCols
	}
	panel := m.mainPanelView(panelW)

	graphBlock := ""
	if m.graphConfig.Enabled && m.view.graphWindow.Open {
		graphBlock = m.renderGraphBlock(panelW)
	}

	drawer := ""
	if m.view.drawerOpen {
		drawer = m.commentDrawerView(panelW)
	}

	parts := make([]string, 0, 6)
	if m.graphConfig.Enabled && m.view.graphWindow.Open {
		parts = append(parts, graphBlock)
	}
	parts = append(parts, panel)
	if m.view.drawerOpen {
		parts = append(parts, drawer)
	}
	parts = append(parts, m.footerView(panelW))
	base := m.styles.App.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
	if m.view.mode != modeTimeWindow || !m.view.timeWindow.Open {
		return base
	}
	return m.renderTimeWindowDialog(base)
}

func (m *Model) mainPanelView(panelWidth int) string {
	if panelWidth < panelMinOuterCols {
		panelWidth = panelMinOuterCols
	}

	innerWidth := panelWidth - 4
	if innerWidth < 1 {
		innerWidth = 1
	}

	innerLines := make([]string, 0, m.viewport.Height+2)
	innerLines = append(innerLines, m.panelHeaderView())
	innerLines = append(innerLines, strings.Repeat("─", innerWidth))
	innerLines = append(innerLines, splitContentLines(m.viewport.View())...)

	panelHeight := m.viewport.Height + panelChromeRows
	return renderBoxedPanel(defaultSaveName(*m), m.currentPanelStatus(), innerLines, panelWidth, panelHeight)
}

func (m *Model) commentDrawerView(panelWidth int) string {
	if panelWidth < panelMinOuterCols {
		panelWidth = panelMinOuterCols
	}
	innerLines := splitContentLines(m.drawerPort.View())
	panelHeight := m.drawerPort.Height + drawerChromeRows
	status := m.currentPanelStatus()
	status.RightText = fmt.Sprintf("Chars %d", m.currentCommentCharCount())
	return renderBoxedPanel("Comment", status, innerLines, panelWidth, panelHeight)
}

func (m *Model) currentPanelStatus() panelStatusSpec {
	totalRows := len(m.table.filteredIndices)
	currentRow := 0
	if totalRows > 0 {
		currentRow = m.cursor + 1
		if currentRow < 1 {
			currentRow = 1
		}
		if currentRow > totalRows {
			currentRow = totalRows
		}
	}

	filterValue := m.filterStatusValue()
	return panelStatusSpec{
		CurrentRow: currentRow,
		TotalRows:  totalRows,
		Filter:     filterValue,
		MarksOn:    m.table.showOnlyMarked,
	}
}

func (m *Model) currentCommentCharCount() int {
	return utf8.RuneCountInString(m.getCommentContent(m.currentRowHashID()))
}

func (m *Model) renderTimeWindowDialog(base string) string {
	dialogW := m.terminalWidth - 12
	if dialogW < 72 {
		dialogW = 72
	}
	if dialogW > 140 {
		dialogW = 140
	}

	body := m.timeWindowDrawerView(dialogW - 4)
	lines := splitContentLines(body)
	box := renderBoxedPanel("Time Window", panelStatusSpec{RightText: "esc: close"}, lines, dialogW, len(lines)+2)

	_ = base // reserved for future backdrop rendering
	return lipgloss.Place(
		m.terminalWidth, m.terminalHeight,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(lipgloss.Color("236")),
	)
}

func (m *Model) metaStatusView(width int) string {
	if width <= 0 {
		return ""
	}

	totalRows := len(m.table.filteredIndices)
	currentRow := 0
	if totalRows > 0 {
		currentRow = m.cursor + 1
		if currentRow < 1 {
			currentRow = 1
		}
		if currentRow > totalRows {
			currentRow = totalRows
		}
	}

	filterValue := m.filterStatusValue()
	filterConfigured := filterValue != "" && !strings.EqualFold(filterValue, "none")

	stateBlock := renderMetaStateBlock(width, currentRow, totalRows, filterValue, filterConfigured, m.table.showOnlyMarked)
	stateWidth := lipgloss.Width(stateBlock)
	leftWidth := width - stateWidth
	if leftWidth <= 0 {
		return stateBlock
	}

	fileName := truncateFilenameMiddlePreserveExt(defaultSaveName(*m), leftWidth)
	fileStyle := lipgloss.NewStyle().
		Bold(false).
		Foreground(lipgloss.Color("252")).
		Width(leftWidth)

	return fileStyle.Render(fileName) + stateBlock
}

type metaField struct {
	label string
	value string
}

func renderMetaStateBlock(maxWidth int, currentRow int, totalRows int, filterValue string, filterActive bool, marksOnly bool) string {
	labelStyle := lipgloss.NewStyle().Bold(false).Faint(true)
	valueStyle := lipgloss.NewStyle().Bold(false)

	includeFilter := filterActive
	includeMarks := marksOnly
	filter := filterValue

	for {
		fields := buildMetaFields(currentRow, totalRows, filter, includeFilter, includeMarks)
		plain := plainMetaFields(fields)
		if lipgloss.Width(plain) <= maxWidth {
			return renderStyledMetaFields(fields, labelStyle, valueStyle)
		}

		if includeFilter && len([]rune(filter)) > 1 {
			filter = truncateEndRunes(filter, len([]rune(filter))-1)
			continue
		}
		if includeMarks {
			includeMarks = false
			continue
		}
		if includeFilter {
			includeFilter = false
			continue
		}

		rowOnly := fmt.Sprintf("Rows %d/%d", currentRow, totalRows)
		if lipgloss.Width(rowOnly) > maxWidth {
			rowOnly = truncateEndRunes(rowOnly, maxWidth)
		}
		if strings.HasPrefix(rowOnly, "Rows ") {
			return labelStyle.Render("Rows") + " " + valueStyle.Render(strings.TrimPrefix(rowOnly, "Rows "))
		}
		return valueStyle.Render(rowOnly)
	}
}

func buildMetaFields(currentRow int, totalRows int, filterValue string, includeFilter bool, includeMarks bool) []metaField {
	fields := []metaField{
		{label: "Rows", value: fmt.Sprintf("%d/%d", currentRow, totalRows)},
	}
	if includeFilter {
		fields = append(fields, metaField{label: "Filter:", value: filterValue})
	}
	if includeMarks {
		fields = append(fields, metaField{label: "Marks:", value: "on"})
	}
	return fields
}
