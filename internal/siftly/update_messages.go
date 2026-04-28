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
	}
	return nil, false
}

func (m *Model) handleDialogInput(msg tea.Msg) (tea.Cmd, bool) {
	if m.activeDialog == nil || !m.activeDialog.IsVisible() {
		return nil, false
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil, false
	}
	logging.Debugf("DIALOG UPDATE: %T got key %q", m.activeDialog, km.String())
	logging.Debugf("model:Update:: Dialog box is active forward update to it")
	var dialogCmd tea.Cmd
	var action dialogs.Action
	m.activeDialog, action, dialogCmd = m.activeDialog.Update(km)
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
	m.activeDialog = dialogs.NewHelpDialog(Keys.Legend(m.graphConfig.Enabled))
	m.activeDialog.Show()
}

func (m *Model) openSaveDialog() {
	logging.Infof("Opening Save dialog")
	m.activeDialog = dialogs.NewSaveDialog(defaultSaveName(*m), defaultDialogDir(*m))
	m.activeDialog.Show()
}

func (m *Model) openExportDialog() {
	logging.Infof("Opening Export dialog")
	m.activeDialog = dialogs.NewExportDialog(defaultExportName(*m), defaultDialogDir(*m))
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
		if err := SaveModel(m, action.Path); err != nil {
			return m.view.notice.Start("Error", "", noticeDuration)
		}
		m.fileName = action.Path
		return m.view.notice.Start("Saved succeeded", "", noticeDuration)
	case dialogs.ActionSaveCancel:
		m.hideActiveDialog()
		return nil
	case dialogs.ActionExportConfirm:
		m.hideActiveDialog()
		if err := ExportModel(m, action.Path); err != nil {
			return m.view.notice.Start("Export Error", "", noticeDuration)
		}
		m.lastExportFileName = action.Path
		return m.view.notice.Start("Exported succeeded", "", noticeDuration)
	case dialogs.ActionExportCancel:
		m.hideActiveDialog()
		return nil
	case dialogs.ActionFilterApply:
		m.hideActiveDialog()
		m.view.command.cmd = CmdFilter
		m.view.mode = modeCommand
		m.view.command.buf = action.Pattern
		m.captureMainBodySnapshot(m.panelWidth())
		return nil
	case dialogs.ActionFilterCancel:
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
