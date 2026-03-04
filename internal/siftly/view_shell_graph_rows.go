package siftly

import (
	"strconv"
	"strings"

	featuregraph "github.com/andareed/siftly-hostlog/internal/siftly/features/graph"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func (m *Model) renderGraphBlock(contentW int) string {
	graphH := m.view.graphWindow.HeightOrDefault()
	graphW := contentW - 2
	if graphW < 0 {
		graphW = 0
	}

	header := make([]string, len(m.table.header))
	for i := range m.table.header {
		header[i] = m.table.header[i].Name
	}
	timeCol, seriesCol, valueCol, ok := featuregraph.ResolveColumnIndices(header, m.graphConfig)
	if !ok {
		return m.styles.GraphArea.Width(contentW).Render(
			ui.RenderGraphMessage(graphW, graphH, "Graph columns not configured"),
		)
	}

	maxKeys := m.view.graphWindow.MaxKeysOrDefault()
	scaleMode := m.graphConfig.ScaleMode
	aggregate := m.graphConfig.Aggregate
	layout := m.graphConfig.Layout
	fillMode := m.graphConfig.FillMode
	cache := m.view.graphCache
	needsPrepare := !cache.valid ||
		cache.dataVersion != m.view.graphDataVersion ||
		cache.rowsLen != len(m.table.rows) ||
		cache.contentW != contentW ||
		cache.graphW != graphW ||
		cache.graphH != graphH ||
		cache.timeCol != timeCol ||
		cache.seriesCol != seriesCol ||
		cache.valueCol != valueCol ||
		cache.maxKeys != maxKeys ||
		cache.scaleMode != scaleMode ||
		cache.aggregate != aggregate ||
		cache.layout != layout ||
		cache.fillMode != fillMode

	if needsPrepare {
		rows := make([][]string, len(m.table.rows))
		for i := range m.table.rows {
			rows[i] = m.table.rows[i].Cols
		}
		prepared := featuregraph.Prepare(featuregraph.Input{
			Width:           graphW,
			Height:          graphH,
			Rows:            rows,
			FilteredIndices: m.table.filteredIndices,
			TimeColumn:      timeCol,
			SeriesColumn:    seriesCol,
			ValueColumn:     valueCol,
			MaxKeys:         maxKeys,
			ScaleMode:       scaleMode,
			AggregateMode:   aggregate,
			LayoutMode:      layout,
			FillMode:        fillMode,
		})
		m.view.graphCache = graphRenderCache{
			valid:       true,
			dataVersion: m.view.graphDataVersion,
			rowsLen:     len(m.table.rows),
			contentW:    contentW,
			graphW:      graphW,
			graphH:      graphH,
			timeCol:     timeCol,
			seriesCol:   seriesCol,
			valueCol:    valueCol,
			maxKeys:     maxKeys,
			scaleMode:   scaleMode,
			aggregate:   aggregate,
			layout:      layout,
			fillMode:    fillMode,
			prepared:    prepared,
		}
	}

	cursorTS := m.graphCursorTimestamp(timeCol)
	content := m.view.graphCache.prepared.Render(cursorTS)
	return m.styles.GraphArea.Width(contentW).Render(content)
}

func (m *Model) graphCursorTimestamp(timeCol int) int64 {
	if m.cursor < 0 || m.cursor >= len(m.table.filteredIndices) {
		return 0
	}
	rowIdx := m.table.filteredIndices[m.cursor]
	if rowIdx < 0 || rowIdx >= len(m.table.rows) {
		return 0
	}
	row := m.table.rows[rowIdx]
	if timeCol < 0 || timeCol >= len(row.Cols) {
		return 0
	}
	ts, err := strconv.ParseInt(strings.TrimSpace(row.Cols[timeCol]), 10, 64)
	if err != nil {
		return 0
	}
	return ts
}

func (m *Model) renderRowAt(filteredIdx int) (string, int, bool) {
	if filteredIdx < 0 || filteredIdx >= len(m.table.filteredIndices) {
		return "", 0, false
	}

	rowIdx := m.table.filteredIndices[filteredIdx]
	row := m.table.rows[rowIdx]
	_, commentPresent := m.table.commentRows[row.ID]

	rendered, height := ui.RenderRow(ui.RowRenderInput{
		Cols:           row.Cols,
		OriginalIndex:  row.OriginalIndex,
		Selected:       filteredIdx == m.cursor,
		SearchQuery:    m.view.searchQuery,
		TotalRows:      len(m.table.rows),
		CommentPresent: commentPresent,
		Mark:           m.table.markedRows[row.ID],
		ColsMeta:       m.table.header,
		Styles: ui.RowStyles{
			Row:             m.styles.Row,
			RowSelected:     m.styles.RowSelected,
			Cell:            m.styles.Cell,
			RedMarker:       m.styles.RedMarker,
			GreenMarker:     m.styles.GreenMarker,
			AmberMarker:     m.styles.AmberMarker,
			SearchHighlight: m.styles.SearchHighlight,
			RowTextFGColor:  m.styles.RowTextFGColor,
			RowSelectedFG:   m.styles.RowSelectedFG,
			RowSelectedBG:   m.styles.RowSelectedBG,
			DefaultMarker:   m.styles.DefaultMarker,
			PillMarker:      m.styles.PillMarker,
			CommentMarker:   m.styles.CommentMarker,
		},
	})
	if m.view.rowHeights == nil {
		m.view.rowHeights = make(map[int]int)
	}
	if cached, ok := m.view.rowHeights[rowIdx]; ok {
		return rendered, cached, true
	}
	m.view.rowHeights[rowIdx] = height
	return rendered, height, true
}
