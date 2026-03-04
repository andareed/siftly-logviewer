package siftly

import (
	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

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

	// Vertical budget:
	// - app frame (margin/padding/border)
	// - panel chrome rows (top border, column header, separator, bottom border)
	// - footer rows
	viewportHeight := height - appV - panelChromeRows - footerRows

	if m.view.drawerOpen {
		m.drawerPort.Width = viewportWidth
		m.drawerPort.Height = drawerContentRows
		m.view.drawerHeight = drawerContentRows + drawerChromeRows
		viewportHeight -= m.view.drawerHeight
	}
	if m.graphConfig.Enabled && m.view.graphWindow.Open {
		graphHeight := m.view.graphWindow.HeightOrDefault()
		viewportHeight -= graphHeight + 2
	}
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	logging.Debugf("Update Received of type Windows Size Message. ViewPort was [%d] and is now getting set to height[%d] width [%d]", m.viewport.Height, viewportHeight, viewportWidth)
	m.viewport.Height = viewportHeight
	m.viewport.Width = viewportWidth
	m.table.header = ui.LayoutColumns(m.table.header, viewportWidth)
}

func (m *Model) refreshView(reason string, withLayout bool) {
	logging.Debugf("refreshView: reason=%s layout=%t", reason, withLayout)
	if withLayout {
		m.recomputeLayout(m.terminalHeight, m.terminalWidth)
	}
	m.clampCursor()
	if m.view.drawerOpen {
		m.refreshDrawerContent()
	}
	m.viewport.SetContent(m.buildViewportContent())
}
