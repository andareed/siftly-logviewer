package siftly

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
	xansi "github.com/charmbracelet/x/ansi"
)

const (
	autoFitColumnMaxWidth  = 64
	autoFitColumnSampleMax = 20_000
)

func (m *Model) openColumnManager() tea.Cmd {
	items := make([]dialogs.ColumnManagerItem, len(m.table.header))
	for i, column := range m.table.header {
		items[i] = dialogs.ColumnManagerItem{
			SourceIndex: column.Index,
			Name:        column.Name,
			Visible:     column.Visible,
			Frozen:      column.Frozen,
			MinWidth:    column.MinWidth,
		}
	}

	logging.Infof("Opening column manager")
	m.activeDialog = dialogs.NewColumnManager(
		items,
		m.table.sortEnabled,
		m.table.sortColumn,
		m.table.sortDesc,
		m.terminalWidth,
		m.terminalHeight,
		m.styles.RowSelectedFG,
		m.styles.RowSelectedBG,
	)
	m.activeDialog.Show()
	return m.activeDialog.Init()
}

func (m *Model) applyColumnManagerResult(result dialogs.ColumnManagerResult) error {
	if len(result.Columns) != len(m.table.header) {
		return fmt.Errorf("column count changed while manager was open")
	}

	bySource := make(map[int]ui.ColumnMeta, len(m.table.header))
	for _, column := range m.table.header {
		if _, duplicate := bySource[column.Index]; duplicate {
			return fmt.Errorf("duplicate source index %d", column.Index)
		}
		bySource[column.Index] = column
	}

	visibleCount := 0
	seen := make(map[int]struct{}, len(result.Columns))
	ordered := make([]ui.ColumnMeta, 0, len(result.Columns))
	for _, item := range result.Columns {
		column, ok := bySource[item.SourceIndex]
		if !ok {
			return fmt.Errorf("column %q is no longer available", item.Name)
		}
		if _, duplicate := seen[item.SourceIndex]; duplicate {
			return fmt.Errorf("column %q appears more than once", item.Name)
		}
		seen[item.SourceIndex] = struct{}{}

		column.Visible = item.Visible
		column.Frozen = item.Visible && item.Frozen
		if item.AutoFit {
			sorted := result.SortEnabled && result.SortColumn == column.Index
			column.MinWidth = m.autoFitColumnWidth(column, sorted)
			column.Weight = 0
		}
		if column.Visible {
			visibleCount++
		}
		ordered = append(ordered, column)
	}
	if visibleCount == 0 {
		return fmt.Errorf("at least one column must remain visible")
	}
	if result.SortEnabled && result.SortColumn != sortColumnOriginalIndex {
		if _, ok := bySource[result.SortColumn]; !ok {
			return fmt.Errorf("sort column is no longer available")
		}
	}

	m.table.header = frozenColumnsFirst(ordered)
	m.table.sortEnabled = result.SortEnabled
	if result.SortEnabled {
		m.table.sortColumn = result.SortColumn
		m.table.sortDesc = result.SortDesc
	} else {
		m.table.sortColumn = -1
		m.table.sortDesc = false
	}

	m.view.columnScrollOffset = 0
	m.table.searchColumns = buildSearchColumnOrder(m.table.header)
	m.rebuildRowOrder()
	m.applyFilter()
	m.recordChange("column layout")
	m.refreshView("column-manager-apply", true)
	return nil
}

func frozenColumnsFirst(columns []ui.ColumnMeta) []ui.ColumnMeta {
	ordered := make([]ui.ColumnMeta, 0, len(columns))
	for _, column := range columns {
		if column.Frozen {
			ordered = append(ordered, column)
		}
	}
	for _, column := range columns {
		if !column.Frozen {
			ordered = append(ordered, column)
		}
	}
	return ordered
}

func (m *Model) autoFitColumnWidth(column ui.ColumnMeta, sorted bool) int {
	width := maxDisplayLineWidth(column.Name)
	if sorted {
		width += 2
	}
	indices := m.table.filteredIndices
	useAllRows := indices == nil
	rowCount := len(indices)
	if useAllRows {
		rowCount = len(m.table.rows)
	}
	step := 1
	if rowCount > autoFitColumnSampleMax {
		step = (rowCount + autoFitColumnSampleMax - 1) / autoFitColumnSampleMax
	}
	lastSampled := -1
	for position := 0; position < rowCount; position += step {
		rowIndex := position
		if !useAllRows {
			rowIndex = indices[position]
		}
		if rowIndex < 0 || rowIndex >= len(m.table.rows) {
			continue
		}
		lastSampled = position
		row := m.table.rows[rowIndex]
		if column.Index < 0 || column.Index >= len(row.Cols) {
			continue
		}
		if rowWidth := maxDisplayLineWidth(row.Cols[column.Index]); rowWidth > width {
			width = rowWidth
		}
		if width >= autoFitColumnMaxWidth {
			return autoFitColumnMaxWidth
		}
	}
	if rowCount > 0 && lastSampled != rowCount-1 {
		rowIndex := rowCount - 1
		if !useAllRows {
			rowIndex = indices[rowCount-1]
		}
		if rowIndex >= 0 && rowIndex < len(m.table.rows) &&
			column.Index >= 0 && column.Index < len(m.table.rows[rowIndex].Cols) {
			if rowWidth := maxDisplayLineWidth(m.table.rows[rowIndex].Cols[column.Index]); rowWidth > width {
				width = rowWidth
			}
		}
	}
	if width < 4 {
		width = 4
	}
	if width > autoFitColumnMaxWidth {
		width = autoFitColumnMaxWidth
	}
	return width
}

func maxDisplayLineWidth(value string) int {
	width := 0
	value = strings.ReplaceAll(value, "\r\n", "\n")
	for _, line := range strings.Split(value, "\n") {
		if lineWidth := xansi.StringWidth(line); lineWidth > width {
			width = lineWidth
		}
	}
	return width
}
