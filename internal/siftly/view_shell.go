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
	return m.renderHeaderLine(m.styles.Header)
}

func (m *Model) renderHeaderLine(headerStyle lipgloss.Style) string {
	cols := make([]ui.HeaderColumn, 0, len(m.table.header))
	for _, col := range m.table.header {
		if !col.Visible || col.Width <= 0 {
			continue
		}
		name := m.sortedHeaderName(col)
		cols = append(cols, ui.HeaderColumn{
			Name:    name,
			Width:   col.Width,
			Visible: col.Visible,
			Frozen:  col.Frozen,
		})
	}

	markerWidth := m.tableMarkerWidth()
	return ui.RenderHeader(
		markerWidth,
		cols,
		m.tableContentWidth(),
		m.view.columnScrollOffset,
		m.styles.Cell,
		headerStyle,
	)
}

func (m *Model) footerView(width int) string {
	logging.Debugf("footerView mode=%d cmd=%d", m.view.mode, m.view.command.cmd)

	footerMode := CmdNone
	prompt := ""
	isInputMode := false
	switch m.view.mode {
	case modeView:
		footerMode = CmdNone
	case modeComment:
		footerMode = CmdComment
		isInputMode = true
		prompt = "comment: "
	case modeCommand:
		isInputMode = true
		footerMode = m.view.command.cmd
		prompt = m.commandPrompt(footerMode) + m.commandValue() + m.commandPreviewSuffix()
	case modeTimeWindow:
		footerMode = CmdTimeWindowSet
		isInputMode = true
		prompt = "edit time window"
	}

	hints := m.footerHints(isInputMode, footerMode)
	selectionCount := m.selectedRowCount()

	debugInfo := ""
	if logging.IsDebugMode() && !isInputMode {
		debugInfo = fmt.Sprintf("dbg term=%dx%d vp=%dx%d cur=%d vis=%d-%d page=%d ch=%d hf=%d abv=%d",
			m.terminalWidth, m.terminalHeight, m.viewport.Width, m.viewport.Height,
			m.cursor, m.view.visibleStart, m.view.visibleEnd, m.pageRowSize,
			m.view.debugCursorHeight, m.view.debugHeightFree, m.view.debugDesiredAboveHeight,
		)
	}

	status := ""
	statusKind := ""
	hintNotice := (m.view.modeHintSeq > 0 && m.view.notice.Seq == m.view.modeHintSeq) ||
		(m.view.prefixHintSeq > 0 && m.view.notice.Seq == m.view.prefixHintSeq)
	if m.view.notice.Msg != "" && !hintNotice {
		status = ui.NoticeText(m.view.notice.Msg, m.view.notice.Type)
		statusKind = m.view.notice.Type
	}
	if status == "" {
		status = m.searchStatusLabel()
	}
	if status == "" && m.table.timeWindow.Enabled {
		status = m.timeWindowStatusLabel()
	}
	if !isInputMode && debugInfo != "" {
		if strings.TrimSpace(status) == "" {
			status = debugInfo
		} else {
			status = status + " | " + debugInfo
		}
	}

	modeBanner := commandLabel(footerMode)
	if m.view.mode == modeTimeWindow {
		modeBanner = "TIME WINDOW"
	}
	if selectionCount > 0 && m.view.mode == modeView {
		modeBanner = "RANGE"
	}
	if m.view.mode == modeView && m.view.inspector.open {
		modeBanner = "INSPECT"
	}
	if m.view.mode == modeView && m.quitConfirmationActive() {
		modeBanner = "QUIT?"
	}
	if m.view.mode == modeView && m.view.pendingViewPrefix != "" {
		switch m.view.pendingViewPrefix {
		case "c":
			modeBanner = "COMMENTS"
		case "t":
			modeBanner = "TIME"
		}
	}
	return ui.RenderFooter(width, ui.FooterState{
		ModeLabel:     modeBanner,
		ActiveStates:  m.footerActiveStates(),
		StatusMessage: status,
		StatusKind:    statusKind,
		Hints:         hints,
		IsInputMode:   isInputMode,
		Prompt:        prompt,
	}, ui.FooterStylesFromTokens(m.styles.ResolvedTokens()))
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
			lipgloss.WithWhitespaceBackground(m.styles.ResolvedTokens().Colors.SurfaceOverlay),
		)
	}

	panelW := m.panelWidth()
	if m.view.mainBodySnapshotActive {
		if !m.mainBodySnapshotValid(panelW) {
			m.captureMainBodySnapshot(panelW)
		}
		base := m.snapshotAppFrameView(panelW)
		if m.view.mode != modeTimeWindow || !m.view.timeWindow.Open {
			return base
		}
		return m.renderTimeWindowDialog(base)
	}

	body := m.mainBodyForView(panelW)
	base := m.appFrameView(body, panelW)
	if m.view.mode != modeTimeWindow || !m.view.timeWindow.Open {
		return base
	}
	return m.renderTimeWindowDialog(base)
}

func (m *Model) mainBodyForView(panelW int) string {
	if m.view.mainBodySnapshotActive {
		if !m.mainBodySnapshotValid(panelW) {
			m.captureMainBodySnapshot(panelW)
		}
		return m.view.mainBodySnapshot
	}
	return m.mainBodyView(panelW)
}

func (m *Model) mainBodySnapshotValid(panelW int) bool {
	return m.view.mainBodySnapshot != "" &&
		m.view.mainBodyFrameSnapshot != "" &&
		m.view.mainBodySnapshotWidth == panelW &&
		m.view.mainBodySnapshotHeight == m.terminalHeight
}

func (m *Model) appFrameView(body string, panelW int) string {
	return m.styles.App.Render(lipgloss.JoinVertical(lipgloss.Left, body, m.footerView(panelW)))
}

func (m *Model) mainBodyView(panelW int) string {
	panel := m.mainPanelView(panelW)

	graphBlock := ""
	if m.graphConfig.Enabled && m.view.graphWindow.Open && m.view.graphHeight > 0 {
		graphBlock = m.renderGraphBlock(panelW)
	}

	drawer := ""
	if m.view.drawerOpen && m.view.drawerHeight > 0 {
		drawer = m.commentDrawerView(panelW)
	}
	inspector := ""
	if m.view.inspector.open && m.view.inspector.height > 0 {
		inspector = m.rowInspectorView(panelW)
	}

	parts := make([]string, 0, 4)
	if m.graphConfig.Enabled && m.view.graphWindow.Open && m.view.graphHeight > 0 {
		parts = append(parts, graphBlock)
	}
	parts = append(parts, panel)
	if m.view.drawerOpen && m.view.drawerHeight > 0 {
		parts = append(parts, drawer)
	}
	if m.view.inspector.open && m.view.inspector.height > 0 {
		parts = append(parts, inspector)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
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
	return renderBoxedPanel(defaultSaveName(*m), m.currentPanelStatus(), innerLines, panelWidth, panelHeight, m.styles.ResolvedTokens())
}

func (m *Model) commentDrawerView(panelWidth int) string {
	if panelWidth < panelMinOuterCols {
		panelWidth = panelMinOuterCols
	}
	innerLines := splitContentLines(m.drawerPort.View())
	panelHeight := m.drawerPort.Height + drawerChromeRows
	status := m.currentPanelStatus()
	status.RightText = fmt.Sprintf("Chars %d", m.currentCommentCharCount())
	return renderBoxedPanel("Comment", status, innerLines, panelWidth, panelHeight, m.styles.ResolvedTokens())
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
		Selected:   m.selectedRowCount(),
	}
}

func (m *Model) currentCommentCharCount() int {
	return utf8.RuneCountInString(m.getCommentContent(m.currentRowHashID()))
}

func (m *Model) renderTimeWindowDialog(base string) string {
	dialogW := responsiveOverlayWidth(m.terminalWidth, 104, 52)

	body := m.timeWindowDrawerView(dialogW - 4)
	lines := splitContentLines(body)
	lines = fitTimeWindowDialogLines(lines, max(1, m.terminalHeight-4))
	box := renderBoxedPanel("Time Window", panelStatusSpec{RightText: "esc: close"}, lines, dialogW, len(lines)+2, m.styles.ResolvedTokens())

	_ = base // reserved for future backdrop rendering
	return lipgloss.Place(
		m.terminalWidth, m.terminalHeight,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(m.styles.ResolvedTokens().Colors.SurfaceOverlay),
	)
}

func responsiveOverlayWidth(terminalWidth, preferred, minimum int) int {
	available := terminalWidth - 4
	if available < 6 {
		available = max(6, terminalWidth)
	}
	if preferred > available {
		preferred = available
	}
	if preferred < minimum && available >= minimum {
		preferred = minimum
	}
	return max(6, preferred)
}

func fitTimeWindowDialogLines(lines []string, maximum int) []string {
	if maximum <= 0 {
		return nil
	}
	result := append([]string(nil), lines...)
	for _, index := range []int{9, 2, 6, 0, 3} {
		if len(result) <= maximum || index >= len(lines) {
			continue
		}
		target := lines[index]
		for i, line := range result {
			if line == target {
				result = append(result[:i], result[i+1:]...)
				break
			}
		}
	}
	if len(result) > maximum {
		result = result[:maximum]
	}
	return result
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

	stateBlock := renderMetaStateBlock(width, currentRow, totalRows, filterValue, filterConfigured, m.table.showOnlyMarked, m.styles.ResolvedTokens())
	stateWidth := lipgloss.Width(stateBlock)
	leftWidth := width - stateWidth
	if leftWidth <= 0 {
		return stateBlock
	}

	fileName := truncateFilenameMiddlePreserveExt(defaultSaveName(*m), leftWidth)
	tokens := m.styles.ResolvedTokens()
	fileStyle := tokens.Emphasis.Strong.Bold(false).Width(leftWidth)

	return fileStyle.Render(fileName) + stateBlock
}

type metaField struct {
	label string
	value string
}

func renderMetaStateBlock(maxWidth int, currentRow int, totalRows int, filterValue string, filterActive bool, marksOnly bool, tokenOptions ...ui.DesignTokens) string {
	tokens := panelDesignTokens(tokenOptions)
	labelStyle := tokens.Emphasis.Muted
	valueStyle := tokens.Emphasis.Normal

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
