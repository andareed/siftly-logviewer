package siftly

import "github.com/andareed/siftly-hostlog/internal/shared/logging"

func (m *Model) panelWidth() int {
	panelW := m.viewport.Width + 4
	if panelW < panelMinOuterCols {
		panelW = panelMinOuterCols
	}
	return panelW
}

func (m *Model) captureMainBodySnapshot(panelW int) {
	if !m.ready {
		return
	}
	m.view.mainBodySnapshot = m.mainBodyView(panelW)
	m.view.mainBodySnapshotWidth = panelW
	m.view.mainBodySnapshotHeight = m.terminalHeight
	m.view.mainBodySnapshotActive = true
	logging.Infof("main body snapshot captured width=%d height=%d mode=%d cmd=%d", panelW, m.terminalHeight, m.view.mode, m.view.command.cmd)
}

func (m *Model) clearMainBodySnapshot() {
	if m.view.mainBodySnapshotActive {
		logging.Infof("main body snapshot released width=%d height=%d mode=%d cmd=%d", m.view.mainBodySnapshotWidth, m.view.mainBodySnapshotHeight, m.view.mode, m.view.command.cmd)
	}
	m.view.mainBodySnapshotActive = false
	m.view.mainBodySnapshot = ""
	m.view.mainBodySnapshotWidth = 0
	m.view.mainBodySnapshotHeight = 0
}
