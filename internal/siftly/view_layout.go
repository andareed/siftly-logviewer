package siftly

import (
	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

const (
	minimumTableRows    = 3
	maximumTableReserve = 8
)

type auxiliaryPanelRequest struct {
	enabled bool
	minimum int
	desired int
}

func (m *Model) recomputeLayout(height int, width int) {
	logging.Debugf("recomputeLayout called with height[%d] width[%d]", height, width)

	appH := m.styles.App.GetHorizontalFrameSize()
	appV := m.styles.App.GetVerticalFrameSize()

	panelWidth := width - appH
	if panelWidth < panelMinOuterCols {
		panelWidth = panelMinOuterCols
	}
	viewportWidth := panelWidth - 4 // "│ " + content + " │"
	if viewportWidth < 1 {
		viewportWidth = 1
	}

	// The table is the primary surface. Auxiliary panels receive their minimum
	// sizes first, then grow together up to their content-aware preferred size.
	bodyHeight := height - appV - panelChromeRows - footerRows
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	m.drawerPort.Width = viewportWidth
	m.inspectorPort.Width = viewportWidth
	drawerContentHeight := m.drawerDesiredContentHeight(viewportWidth)
	inspectorContentHeight := m.inspectorDesiredContentHeight(viewportWidth)
	graphContentHeight := m.responsiveGraphContentHeight(bodyHeight)
	graphChromeRows := m.styles.GraphArea.GetVerticalFrameSize()

	requests := []auxiliaryPanelRequest{
		{
			enabled: m.view.drawerOpen,
			minimum: drawerChromeRows + 1,
			desired: drawerChromeRows + drawerContentHeight,
		},
		{
			enabled: m.view.inspector.open,
			minimum: inspectorChromeRows + 1,
			desired: inspectorChromeRows + inspectorContentHeight,
		},
		{
			enabled: m.graphConfig.Enabled && m.view.graphWindow.Open,
			minimum: graphChromeRows + 1,
			desired: graphChromeRows + graphContentHeight,
		},
	}
	allocations := allocateAuxiliaryPanels(bodyHeight, requests)

	m.view.drawerHeight = allocations[0]
	if m.view.drawerHeight >= drawerChromeRows+1 {
		m.drawerPort.Height = m.view.drawerHeight - drawerChromeRows
	} else {
		m.view.drawerHeight = 0
		m.drawerPort.Height = 0
	}

	m.view.inspector.height = allocations[1]
	if m.view.inspector.height >= inspectorChromeRows+1 {
		m.inspectorPort.Height = m.view.inspector.height - inspectorChromeRows
	} else {
		m.view.inspector.height = 0
		m.inspectorPort.Height = 0
	}

	graphOuterHeight := allocations[2]
	m.view.graphHeight = graphOuterHeight - graphChromeRows
	if m.view.graphHeight < 1 {
		m.view.graphHeight = 0
		graphOuterHeight = 0
	}

	viewportHeight := bodyHeight - m.view.drawerHeight - m.view.inspector.height - graphOuterHeight
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	logging.Debugf("Update Received of type Windows Size Message. ViewPort was [%d] and is now getting set to height[%d] width [%d]", m.viewport.Height, viewportHeight, viewportWidth)
	m.viewport.Height = viewportHeight
	m.viewport.Width = viewportWidth
	m.viewport.SetXOffset(0)
	m.table.header = ui.LayoutColumns(m.table.header, max(1, viewportWidth-m.tableMarkerWidth()))
	m.view.rowHeights = nil
	m.clampColumnScrollOffset()
}

func allocateAuxiliaryPanels(bodyHeight int, requests []auxiliaryPanelRequest) []int {
	allocations := make([]int, len(requests))
	if bodyHeight <= 1 {
		return allocations
	}

	tableReserve := bodyHeight / 4
	if tableReserve < minimumTableRows {
		tableReserve = minimumTableRows
	}
	if tableReserve > maximumTableReserve {
		tableReserve = maximumTableReserve
	}
	if tableReserve >= bodyHeight {
		tableReserve = bodyHeight - 1
	}
	budget := bodyHeight - tableReserve

	minimumTotal := 0
	for _, request := range requests {
		if request.enabled {
			minimumTotal += request.minimum
		}
	}
	if minimumTotal > budget {
		budget = bodyHeight - 1
	}

	remaining := budget
	for i, request := range requests {
		if !request.enabled || request.minimum <= 0 || remaining < request.minimum {
			continue
		}
		allocations[i] = request.minimum
		remaining -= request.minimum
	}

	for remaining > 0 {
		grew := false
		for i, request := range requests {
			if !request.enabled || allocations[i] == 0 || allocations[i] >= request.desired {
				continue
			}
			allocations[i]++
			remaining--
			grew = true
			if remaining == 0 {
				break
			}
		}
		if !grew {
			break
		}
	}
	return allocations
}

func (m *Model) responsiveGraphContentHeight(bodyHeight int) int {
	desired := m.view.graphWindow.HeightOrDefault()
	responsiveMaximum := bodyHeight / 2
	if responsiveMaximum < 3 {
		responsiveMaximum = 3
	}
	if desired > responsiveMaximum {
		desired = responsiveMaximum
	}
	if desired < 1 {
		desired = 1
	}
	return desired
}

func (m *Model) refreshView(reason string, withLayout bool) {
	logging.Debugf("refreshView: reason=%s layout=%t", reason, withLayout)
	m.clampCursor()
	if withLayout {
		m.recomputeLayout(m.terminalHeight, m.terminalWidth)
	} else if m.view.inspector.open || m.view.drawerOpen {
		m.recomputeLayout(m.terminalHeight, m.terminalWidth)
	}
	if m.view.drawerOpen {
		m.refreshDrawerContent()
	}
	if m.view.inspector.open {
		m.refreshInspectorContent()
	}
	m.viewport.SetContent(m.buildViewportContent())
}
