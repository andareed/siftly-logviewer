package siftly

import (
	"fmt"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleSystemMsg(msg tea.Msg) (tea.Cmd, bool) {
	switch fmt.Sprintf("%T", msg) {
	case "cursor.BlinkMsg", "cursor.BlinkCanceledMsg":
		return nil, true
	}
	switch msg := msg.(type) {
	case ui.ClearNoticeMsg:
		m.view.notice.ApplyClear(msg)
		return nil, true
	case operationTickMsg:
		return m.handleOperationTick(msg), true
	case searchIndexChunkMsg:
		return m.handleSearchIndexChunk(msg), true
	case saveCompleteMsg:
		return m.handleSaveComplete(msg), true
	case graphExportCompleteMsg:
		return m.handleGraphExportComplete(msg), true
	case filterCompleteMsg:
		return m.handleFilterComplete(msg), true
	case fullSourceReloadCompleteMsg:
		return m.handleFullSourceReloadComplete(msg), true
	case recoveryFlushMsg:
		return m.handleRecoveryFlush(msg), true
	case recoveryWriteCompleteMsg:
		return m.handleRecoveryWriteComplete(msg), true
	}
	return nil, false
}

func (m *Model) handleDialogInput(msg tea.Msg) (tea.Cmd, bool) {
	if m.activeDialog == nil || !m.activeDialog.IsVisible() {
		return nil, false
	}
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		return nil, false
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		logging.Debugf("DIALOG UPDATE: %T got key %q", m.activeDialog, km.String())
	}
	logging.Debugf("model:Update:: Dialog box is active forward update to it")
	var dialogCmd tea.Cmd
	var action dialogs.Action
	m.activeDialog, action, dialogCmd = m.activeDialog.Update(msg)
	logging.Debugf("DIALOG ACTION: kind=%d command=%q", action.Kind, action.CommandID)
	actionCmd := m.applyDialogAction(action)
	return batchCmd(dialogCmd, actionCmd), true
}

func (m *Model) handleWindowMsg(msg tea.Msg) (tea.Cmd, bool) {
	win, ok := msg.(tea.WindowSizeMsg)
	if !ok {
		return nil, false
	}
	m.terminalHeight = win.Height
	m.terminalWidth = win.Width
	if dialog, ok := m.activeDialog.(dialogs.Resizable); ok && m.activeDialog.IsVisible() {
		dialog.Resize(win.Width, win.Height)
	}
	m.viewport = viewport.New(0, 0) // TODO: Pretty sure this is redundant
	m.ready = true
	m.refreshView("window-size", true)
	if m.view.mainBodySnapshotActive {
		m.captureMainBodySnapshot(m.panelWidth())
	}
	return nil, true
}

func (m *Model) openHelpDialog() {
	logging.Infof("Opening Help dialog")
	m.activeDialog = dialogs.NewHelpDialog(m.commandItems(), m.terminalWidth, m.terminalHeight, m.styles.ResolvedTokens())
	m.activeDialog.Show()
}

func (m *Model) openCommandPalette() tea.Cmd {
	logging.Infof("Opening command palette")
	m.activeDialog = dialogs.NewCommandPalette(
		m.commandItems(),
		m.terminalWidth,
		m.terminalHeight,
		m.styles.RowSelectedFG,
		m.styles.RowSelectedBG,
		m.styles.ResolvedTokens(),
	)
	m.activeDialog.Show()
	return m.activeDialog.Init()
}

func (m *Model) openSaveDialog() {
	logging.Infof("Opening Save dialog")
	dialog := dialogs.NewSaveDialog(defaultSaveName(*m), defaultDialogDir(*m), m.styles.ResolvedTokens())
	dialog.Resize(m.terminalWidth, m.terminalHeight)
	m.activeDialog = dialog
	m.activeDialog.Show()
}

func (m *Model) openExportDialog() {
	logging.Infof("Opening filtered data export dialog")
	dialog := dialogs.NewExportDialog(defaultExportName(*m), defaultDialogDir(*m), m.styles.ResolvedTokens())
	dialog.Resize(m.terminalWidth, m.terminalHeight)
	m.activeDialog = dialog
	m.activeDialog.Show()
}

func (m *Model) hideActiveDialog() {
	if m.activeDialog != nil {
		m.activeDialog.Hide()
	}
}

func (m *Model) applyDialogAction(action dialogs.Action) tea.Cmd {
	switch action.Kind {
	case dialogs.ActionNone:
		return nil
	case dialogs.ActionClose:
		m.hideActiveDialog()
		return nil
	case dialogs.ActionSaveConfirm:
		m.hideActiveDialog()
		return m.startSaveOperation(action.Path)
	case dialogs.ActionSaveCancel:
		m.hideActiveDialog()
		return nil
	case dialogs.ActionExportConfirm:
		m.hideActiveDialog()
		if err := ExportModel(m, action.Path); err != nil {
			return m.view.notice.Start("Filtered data export error", "error", noticeDuration)
		}
		m.lastExportFileName = action.Path
		return m.view.notice.Start("Filtered data exported", "success", noticeDuration)
	case dialogs.ActionExportCancel:
		m.hideActiveDialog()
		return nil
	case dialogs.ActionFilterApply:
		m.hideActiveDialog()
		m.view.command = newCommandInput(CmdFilter, action.Pattern)
		m.view.mode = modeCommand
		m.captureMainBodySnapshot(m.panelWidth())
		return nil
	case dialogs.ActionFilterCancel:
		m.hideActiveDialog()
		return nil
	case dialogs.ActionCommandRun:
		m.hideActiveDialog()
		return m.runPaletteCommand(action.CommandID)
	case dialogs.ActionCommandCancel:
		m.hideActiveDialog()
		return nil
	case dialogs.ActionColumnManagerApply:
		m.hideActiveDialog()
		if action.ColumnManager == nil {
			return m.view.notice.Start("Column layout was not applied", "error", noticeDuration)
		}
		if err := m.applyColumnManagerResult(*action.ColumnManager); err != nil {
			return m.view.notice.Start(err.Error(), "error", noticeDuration)
		}
		return m.view.notice.Start("Column layout applied", "success", noticeDuration)
	case dialogs.ActionColumnManagerCancel:
		m.hideActiveDialog()
		return nil
	default:
		return nil
	}
}

func batchCmd(left, right tea.Cmd) tea.Cmd {
	switch {
	case left == nil:
		return right
	case right == nil:
		return left
	default:
		return tea.Batch(left, right)
	}
}
