package siftly

import (
	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	modeView mode = iota
	// modeFilter
	// modeMarking
	modeComment
	modeCommand
	modeTimeWindow
)

//TODO Replace the name renderedRow, as these are not rendered anymore

type Model struct {
	viewport            viewport.Model
	drawerPort          viewport.Model
	ready               bool
	cursor              int // index into rows
	lastVisibleRowCount int
	terminalHeight      int
	terminalWidth       int
	pageRowSize         int
	activeDialog        dialogs.Dialog
	filterConfig        FilterConfigSettings
	fileName            string // filename the data will be saved to
	InitialPath         string
	lastExportFileName  string
	view                viewState
	table               tableState
	styles              ui.Styles
	graphConfig         GraphConfig
}

// SetStyles injects UI styles from the wrapper package (e.g., hostlog).
func (m *Model) SetStyles(styles ui.Styles) {
	m.styles = styles
}

// SetGraphConfig enables optional graph rendering for callers that provide graph semantics.
func (m *Model) SetGraphConfig(cfg GraphConfig) {
	m.graphConfig = cfg
	if !cfg.Enabled {
		m.view.graphWindow.Open = false
		return
	}
	m.view.graphWindow.MaxKeys = cfg.MaxKeys
	m.view.graphWindow.Height = cfg.Height
}

func (m *Model) InitialiseView() {
	m.ensureTableDerivedState()
	m.table.showOnlyMarked = false
	m.drawerPort = viewport.New(0, 0)
	m.view.drawerHeight = 13 // TODO:should be a better way of calcing this rather than hardcoding
	m.view.drawerOpen = false
	m.view.mode = modeView
	if m.graphConfig.Enabled {
		m.view.graphWindow.MaxKeys = m.graphConfig.MaxKeys
		m.view.graphWindow.Height = m.graphConfig.Height
	}
	m.initTimeWindowState()
	if m.table.timeWindow.Enabled {
		m.applyFilter()
	}
}

func (m *Model) Init() tea.Cmd {
	defer logging.TimeOperation("initial filter")()
	m.applyFilter()
	logging.Info("siftly-hostlog: Initialised")
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		logging.Debugf("KEY: %q type=%v mode=%v dialog=%T visible=%v",
			km.String(), km.Type, m.view.mode,
			m.activeDialog, m.activeDialog != nil && m.activeDialog.IsVisible(),
		)
	}

	logging.Debugf("model:Update called with msg: %#T: %#v", msg, msg)

	if cmd, handled := m.handleSystemMsg(msg); handled {
		return m, cmd
	}
	if cmd, handled := m.handleDialogInput(msg); handled {
		return m, cmd
	}
	if cmd, handled := m.handleWindowMsg(msg); handled {
		return m, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeyMsg(keyMsg)
	}
	return m, nil
}
