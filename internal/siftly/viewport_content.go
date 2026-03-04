package siftly

import "github.com/andareed/siftly-hostlog/internal/shared/logging"

type viewportRowsResult struct {
	rows               []string
	startIdx           int
	endIdx             int
	cursorHeight       int
	heightFree         int
	desiredAboveHeight int
}

func (m *Model) buildViewportContent() string {
	logging.Debug("buildViewportContent called")
	viewportHeight := m.viewport.Height

	cursor := m.cursor

	if len(m.table.filteredIndices) == 0 && cursor < 0 {
		logging.Debugf("renderTable: Returning blank filteredIndices Lenght[%d] cursor[%d]", len(m.table.filteredIndices), cursor)

		return ""
	}
	//TODO: Defect here as we should be using the row count not the display index to maintain between a filter and non-filtered list
	if len(m.table.filteredIndices) < cursor {
		m.cursor = 0
		cursor = 0
	}
	content, result, ok := m.composeViewportContent(cursor, viewportHeight)
	if !ok {
		return ""
	}
	m.view.visibleStart = result.startIdx
	m.view.visibleEnd = result.endIdx
	// Metrics
	m.pageRowSize = len(result.rows)
	m.lastVisibleRowCount = len(result.rows)
	return content
}

func (m *Model) computeVisibleRows(cursor int, viewportHeight int) ([]string, int, int) {
	// Compute rows that fit in the viewport around the cursor, preserving visual balance.
	result, ok := m.collectViewportRows(cursor, viewportHeight)
	if !ok {
		return nil, 0, 0
	}
	// Capture diagnostics for debug overlays.
	m.view.debugCursorHeight = result.cursorHeight
	m.view.debugHeightFree = result.heightFree
	m.view.debugDesiredAboveHeight = result.desiredAboveHeight
	return result.rows, result.startIdx, result.endIdx
}

func (m *Model) collectViewportRows(cursor int, viewportHeight int) (viewportRowsResult, bool) {
	rowCount := len(m.table.filteredIndices)
	cursorRenderedRow, cursorHeight, ok := m.renderRowAt(cursor)
	if !ok {
		return viewportRowsResult{}, false
	}

	heightFree := viewportHeight - cursorHeight
	desiredAboveHeight := heightFree / 2
	if desiredAboveHeight < 0 {
		desiredAboveHeight = 0
	}

	upIndex := cursor - 1
	downIndex := cursor + 1

	var above []string
	var below []string

	aboveHeight := 0
	for heightFree > 0 && (upIndex >= 0 || downIndex < rowCount) {
		if upIndex >= 0 && aboveHeight < desiredAboveHeight {
			rendered, height, ok := m.renderRowAt(upIndex)
			if ok && height <= heightFree {
				above = append(above, rendered)
				heightFree -= height
				aboveHeight += height
				upIndex--
				continue
			}
		}
		if downIndex < rowCount {
			rendered, height, ok := m.renderRowAt(downIndex)
			if ok && height <= heightFree {
				below = append(below, rendered)
				heightFree -= height
				downIndex++
				continue
			}
		}
		if upIndex >= 0 {
			rendered, height, ok := m.renderRowAt(upIndex)
			if ok && height <= heightFree {
				above = append(above, rendered)
				heightFree -= height
				aboveHeight += height
				upIndex--
				continue
			}
		}
		break
	}

	renderedRows := make([]string, 0, len(above)+1+len(below))
	for i := len(above) - 1; i >= 0; i-- {
		renderedRows = append(renderedRows, above[i])
	}
	renderedRows = append(renderedRows, cursorRenderedRow)
	renderedRows = append(renderedRows, below...)

	startIdx := cursor - len(above)
	endIdx := cursor + len(below)
	return viewportRowsResult{
		rows:               renderedRows,
		startIdx:           startIdx,
		endIdx:             endIdx,
		cursorHeight:       cursorHeight,
		heightFree:         heightFree,
		desiredAboveHeight: desiredAboveHeight,
	}, true
}

func (m *Model) composeViewportContent(cursor int, viewportHeight int) (string, viewportRowsResult, bool) {
	result, ok := m.collectViewportRows(cursor, viewportHeight)
	if !ok {
		return "", viewportRowsResult{}, false
	}
	if len(result.rows) == 0 {
		return "", result, true
	}
	totalLen := 0
	for _, row := range result.rows {
		totalLen += len(row) + 1
	}
	b := make([]byte, 0, totalLen)
	for _, row := range result.rows {
		b = append(b, row...)
		b = append(b, '\n')
	}
	return string(b), result, true
}
