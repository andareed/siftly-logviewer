package siftly

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	xansi "github.com/charmbracelet/x/ansi"
)

const rowInspectorTitle = "Row Inspector"

func (m *Model) rowInspectorView(panelWidth int) string {
	if panelWidth < panelMinOuterCols {
		panelWidth = panelMinOuterCols
	}

	status := panelStatusSpec{RightText: "No row"}
	if row, _, ok := m.currentInspectorRow(); ok {
		fieldPosition := 0
		if len(m.table.header) > 0 {
			fieldPosition = m.inspectorFieldIndex() + 1
		}
		status.RightText = m.inspectorStatusText(panelWidth, row, fieldPosition)
	}

	innerLines := splitContentLines(m.inspectorPort.View())
	panelHeight := m.inspectorPort.Height + inspectorChromeRows
	return renderBoxedPanel(rowInspectorTitle, status, innerLines, panelWidth, panelHeight, m.styles.ResolvedTokens())
}

func (m *Model) inspectorStatusText(panelWidth int, row Row, fieldPosition int) string {
	mark := strings.ToUpper(string(m.table.markedRows[row.ID]))
	if mark == "" {
		mark = "UNMARKED"
	}
	current := m.cursor + 1
	total := len(m.table.filteredIndices)
	fieldCount := len(m.table.header)
	candidates := []string{
		fmt.Sprintf("Row %d/%d  Source %d  Field %d/%d  %s", current, total, row.OriginalIndex, fieldPosition, fieldCount, mark),
		fmt.Sprintf("%d/%d  Src %d  F %d/%d  %s", current, total, row.OriginalIndex, fieldPosition, fieldCount, mark),
		fmt.Sprintf("%d/%d  F %d/%d", current, total, fieldPosition, fieldCount),
	}
	budget := panelWidth - xansi.StringWidth(rowInspectorTitle) - 8
	for _, candidate := range candidates {
		if xansi.StringWidth(candidate) <= budget {
			return candidate
		}
	}
	return candidates[len(candidates)-1]
}

func (m *Model) refreshInspectorContent() {
	row, rowIndex, ok := m.currentInspectorRow()
	if !ok {
		m.inspectorPort.SetContent("No row selected")
		m.inspectorPort.SetYOffset(0)
		m.view.inspector.hasContent = false
		m.view.inspector.lastRowIndex = -1
		return
	}

	field := m.inspectorFieldIndex()
	content, fieldOffsets := m.buildInspectorContent(row, field, m.inspectorPort.Width)
	rowChanged := !m.view.inspector.hasContent || m.view.inspector.lastRowIndex != rowIndex
	fieldChanged := !m.view.inspector.hasContent || m.view.inspector.lastField != field
	widthChanged := !m.view.inspector.hasContent || m.view.inspector.lastWidth != m.inspectorPort.Width

	m.inspectorPort.SetContent(content)
	if rowChanged || fieldChanged || widthChanged {
		targetOffset := 0
		if field > 0 || (!rowChanged && fieldChanged) {
			targetOffset = inspectorFieldOffset(fieldOffsets, field)
		}
		m.inspectorPort.SetYOffset(targetOffset)
	}

	m.view.inspector.lastRowIndex = rowIndex
	m.view.inspector.lastField = field
	m.view.inspector.lastWidth = m.inspectorPort.Width
	m.view.inspector.hasContent = true
}

func (m *Model) buildInspectorContent(row Row, selectedField int, width int) (string, []int) {
	if width < 1 {
		width = 1
	}

	lines := make([]string, 0, len(m.table.header)+2)
	tokens := m.styles.ResolvedTokens()
	fieldOffsets := make([]int, len(m.table.header))
	gridColumns := 1
	const gridGap = 3
	cellWidth := width
	if width >= 84 {
		gridColumns = 2
		cellWidth = (width - gridGap) / gridColumns
	}

	type compactField struct {
		index int
		label string
		value string
	}
	pending := make([]compactField, 0, gridColumns)
	flushPending := func() {
		if len(pending) == 0 {
			return
		}
		rowOffset := len(lines)
		cells := make([]string, len(pending))
		for i, field := range pending {
			fieldOffsets[field.index] = rowOffset
			cells[i] = renderCompactInspectorField(field.label, field.value, field.index == selectedField, tokens)
			if i < len(pending)-1 {
				cells[i] = padInspectorCell(cells[i], cellWidth)
			}
		}
		lines = append(lines, strings.Join(cells, strings.Repeat(" ", gridGap)))
		pending = pending[:0]
	}

	for fieldIndex, column := range m.table.header {
		label := inspectorColumnLabel(column, fieldIndex)
		value := normalizeInspectorText(inspectorColumnValue(row, column, fieldIndex))
		if value == "" {
			value = "(empty)"
		}

		compactWidth := xansi.StringWidth("  " + label + "  " + value)
		if gridColumns > 1 && !strings.Contains(value, "\n") && compactWidth <= cellWidth {
			pending = append(pending, compactField{index: fieldIndex, label: label, value: value})
			if len(pending) == gridColumns {
				flushPending()
			}
			continue
		}

		flushPending()
		fieldOffsets[fieldIndex] = len(lines)
		lines = append(lines, renderFullWidthInspectorField(label, value, width, fieldIndex == selectedField, tokens)...)
	}
	flushPending()

	if len(m.table.header) == 0 {
		lines = append(lines, "No columns")
	}
	if comment := normalizeInspectorText(m.table.commentRows[row.ID]); comment != "" {
		lines = append(lines, renderFullWidthInspectorField("Comment", comment, width, false, tokens)...)
	}
	return strings.Join(lines, "\n"), fieldOffsets
}

func (m *Model) inspectorDesiredContentHeight(width int) int {
	row, _, ok := m.currentInspectorRow()
	if !ok {
		return 1
	}
	content, _ := m.buildInspectorContent(row, m.inspectorFieldIndex(), width)
	height := len(splitContentLines(content))
	if height < 1 {
		height = 1
	}
	if height > inspectorMaxContentRows {
		height = inspectorMaxContentRows
	}
	return height
}

func renderCompactInspectorField(label, value string, selected bool, tokenOptions ...ui.DesignTokens) string {
	return renderInspectorLabel(label, selected, tokenOptions...) + "  " + value
}

func renderFullWidthInspectorField(label, value string, width int, selected bool, tokenOptions ...ui.DesignTokens) []string {
	if width < 1 {
		width = 1
	}
	labelWidth := xansi.StringWidth("  " + label)
	if labelWidth > width-1 {
		labelLines := wrapInspectorText(label, width-2)
		lines := make([]string, 0, len(labelLines)+1)
		for i, line := range labelLines {
			lines = append(lines, renderInspectorLabelPart(line, selected, i == 0, tokenOptions...))
		}
		return append(lines, indentInspectorText(value, width, 4)...)
	}

	valueWidth := width - labelWidth - 2
	if valueWidth < 12 {
		return append([]string{renderInspectorLabel(label, selected, tokenOptions...)}, indentInspectorText(value, width, 4)...)
	}

	valueLines := wrapInspectorText(value, valueWidth)
	lines := make([]string, 0, len(valueLines))
	lines = append(lines, renderInspectorLabel(label, selected, tokenOptions...)+"  "+valueLines[0])
	continuationIndent := strings.Repeat(" ", labelWidth+2)
	for _, line := range valueLines[1:] {
		lines = append(lines, continuationIndent+line)
	}
	return lines
}

func renderInspectorLabel(label string, selected bool, tokenOptions ...ui.DesignTokens) string {
	return renderInspectorLabelPart(label, selected, true, tokenOptions...)
}

func renderInspectorLabelPart(label string, selected, firstLine bool, tokenOptions ...ui.DesignTokens) string {
	prefix := "  "
	if selected && firstLine {
		prefix = "› "
	}
	tokens := panelDesignTokens(tokenOptions)
	style := tokens.Emphasis.Strong
	if selected {
		style = tokens.States.Selected.Bold(true)
	}
	return style.Render(prefix + label)
}

func padInspectorCell(cell string, width int) string {
	padding := width - xansi.StringWidth(cell)
	if padding < 0 {
		padding = 0
	}
	return cell + strings.Repeat(" ", padding)
}

func (m *Model) cycleInspectorField(delta int) {
	count := len(m.table.header)
	if count == 0 {
		m.view.inspector.selectedField = 0
		return
	}
	field := (m.inspectorFieldIndex() + delta) % count
	if field < 0 {
		field += count
	}
	m.view.inspector.selectedField = field
}

func (m *Model) inspectorFieldIndex() int {
	count := len(m.table.header)
	if count == 0 {
		m.view.inspector.selectedField = 0
		return 0
	}
	if m.view.inspector.selectedField < 0 {
		m.view.inspector.selectedField = 0
	}
	if m.view.inspector.selectedField >= count {
		m.view.inspector.selectedField = count - 1
	}
	return m.view.inspector.selectedField
}

func (m *Model) currentInspectorRow() (Row, int, bool) {
	if m.cursor < 0 || m.cursor >= len(m.table.filteredIndices) {
		return Row{}, -1, false
	}
	rowIndex := m.table.filteredIndices[m.cursor]
	if rowIndex < 0 || rowIndex >= len(m.table.rows) {
		return Row{}, -1, false
	}
	return m.table.rows[rowIndex], rowIndex, true
}

func inspectorColumnValue(row Row, column ui.ColumnMeta, fieldIndex int) string {
	sourceIndex := column.Index
	if sourceIndex < 0 || sourceIndex >= len(row.Cols) {
		sourceIndex = fieldIndex
	}
	if sourceIndex < 0 || sourceIndex >= len(row.Cols) {
		return ""
	}
	return row.Cols[sourceIndex]
}

func inspectorColumnLabel(column ui.ColumnMeta, fieldIndex int) string {
	if strings.TrimSpace(column.Name) == "" {
		return fmt.Sprintf("Column %d", fieldIndex+1)
	}
	return column.Name
}

func indentInspectorText(value string, width, indent int) []string {
	if indent < 0 {
		indent = 0
	}
	contentWidth := width - indent
	if contentWidth < 1 {
		contentWidth = 1
	}
	prefix := strings.Repeat(" ", indent)
	wrapped := wrapInspectorText(value, contentWidth)
	lines := make([]string, len(wrapped))
	for i, line := range wrapped {
		lines[i] = prefix + line
	}
	return lines
}

func wrapInspectorText(value string, width int) []string {
	value = normalizeInspectorText(value)
	if value == "" {
		return []string{""}
	}
	if width < 1 {
		width = 1
	}
	return strings.Split(xansi.Wrap(value, width, ""), "\n")
}

func normalizeInspectorText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\t", "    ")
}

func inspectorFieldOffset(offsets []int, field int) int {
	if field < 0 || field >= len(offsets) {
		return 0
	}
	return offsets[field]
}
