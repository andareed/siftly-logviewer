package siftly

import (
	"fmt"
	"strings"
	"time"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type viewAction int

const (
	viewActionNone viewAction = iota
	viewActionOpenColumnManager
	viewActionPrefixComment
	viewActionPrefixTime
	viewActionPrefixExport
	viewActionJump
	viewActionFilter
	viewActionSearch
	viewActionMarkMode
	viewActionQuit
	viewActionCancelQuit
	viewActionUndo
	viewActionRedo
	viewActionCopyRow
	viewActionToggleInspector
	viewActionInspectorNextField
	viewActionInspectorPrevField
	viewActionInspectorCopyField
	viewActionInspectorScrollDown
	viewActionInspectorScrollUp
	viewActionToggleRange
	viewActionClearRange
	viewActionJumpToStart
	viewActionJumpToEnd
	viewActionToggleShowMarks
	viewActionNextMark
	viewActionPrevMark
	viewActionSearchNext
	viewActionSearchPrev
	viewActionToggleFilter
	viewActionToggleGraph
	viewActionRowDown
	viewActionRowUp
	viewActionPageUp
	viewActionPageDown
	viewActionOpenPalette
	viewActionOpenHelp
	viewActionScrollLeft
	viewActionScrollRight
	viewActionSave
)

const (
	quitConfirmWindow = 10 * time.Second
)

type viewPrefixAction int

const (
	viewPrefixActionNone viewPrefixAction = iota
	viewPrefixActionCommentEdit
	viewPrefixActionCommentToggleDrawer
	viewPrefixActionTimeWindowOpen
	viewPrefixActionTimeSetStart
	viewPrefixActionTimeSetEnd
	viewPrefixActionTimeReset
	viewPrefixActionExportData
	viewPrefixActionExportGraph
	viewPrefixActionCancel
)

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// if m.activeDialog != nil && m.activeDialog.IsVisible() {
	//
	// return m.handleDialogKey(msg)
	// }
	switch m.view.mode {
	case modeView:
		return m.handleViewModeKey(msg)
	case modeCommand:
		return m.handleCommandKey(msg)
	case modeTimeWindow:
		return m.handleTimeWindowKey(msg)
	}

	return m, nil
}

func (m *Model) handleViewModeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	didRefresh := false

	if handled, comboCmd, refresh := m.handleViewPrefixKey(msg); handled {
		if refresh {
			m.refreshView("view-prefix", true)
			return m, comboCmd
		}
		if m.ready && !didRefresh {
			m.refreshView("view-key", false)
		}
		return m, comboCmd
	}

	action := m.resolveViewAction(msg)
	if action != viewActionQuit && action != viewActionCancelQuit {
		m.view.quitConfirmUntil = time.Time{}
	}
	switch action {
	case viewActionOpenColumnManager:
		return m, m.openColumnManager()
	case viewActionPrefixComment:
		m.view.pendingViewPrefix = "c"
		m.setPrefixHint("e: edit comment   v: toggle drawer   esc: cancel")
		cmd = nil
		return m, cmd
	case viewActionPrefixTime:
		m.view.pendingViewPrefix = "t"
		m.setPrefixHint("w: window   b: set start   e: set end   r: reset   esc: cancel")
		cmd = nil
		return m, cmd
	case viewActionPrefixExport:
		m.view.pendingViewPrefix = "e"
		if m.graphConfig.Enabled {
			m.setPrefixHint("d: filtered data   g: graph SVG   esc: cancel")
		} else {
			m.setPrefixHint("d: filtered data   esc: cancel")
		}
		return m, nil
	// Migrating to a command / input method
	case viewActionJump:
		logging.Infof("Enabling Command: Jumping to specific line number if it exists")
		cmd = m.enterCommand(CmdJump, "", true, false)
	case viewActionFilter:
		logging.Infof("Enabling Command: Filtering")
		cmd = m.enterCommand(CmdFilter, "", true, false)
	case viewActionSearch:
		logging.Infof("Enabling Command: Search")
		cmd = m.enterCommand(CmdSearch, "", true, false)
	case viewActionMarkMode:
		logging.Infof("Enable COmmand: Marking")
		cmd = m.enterCommand(CmdMark, "", true, false)
	//TODO: Implement Serach
	case viewActionQuit:
		return m, m.confirmOrQuit()
	case viewActionCancelQuit:
		m.view.quitConfirmUntil = time.Time{}
		cmd = m.view.notice.Start("Quit cancelled", "", noticeDuration)
	case viewActionUndo:
		cmd = m.undoLastChange()
		didRefresh = true
	case viewActionRedo:
		cmd = m.redoLastChange()
		didRefresh = true
	case viewActionCopyRow:
		logging.Infof("Key combination for copying rows to the clipboard")
		cmd = m.copyRowsToClipboard()
	case viewActionToggleInspector:
		m.view.inspector.open = !m.view.inspector.open
		if m.view.inspector.open {
			m.view.drawerOpen = false
			m.view.inspector.hasContent = false
		}
		m.refreshView("inspector-toggle", true)
		didRefresh = true
	case viewActionInspectorNextField:
		m.cycleInspectorField(1)
	case viewActionInspectorPrevField:
		m.cycleInspectorField(-1)
	case viewActionInspectorCopyField:
		cmd = m.copyInspectorFieldToClipboard()
	case viewActionInspectorScrollDown:
		m.inspectorPort.ScrollDown(1)
		didRefresh = true
	case viewActionInspectorScrollUp:
		m.inspectorPort.ScrollUp(1)
		didRefresh = true
	case viewActionToggleRange:
		m.toggleRowRangeSelection()
	case viewActionClearRange:
		m.clearRowRangeSelection()
	case viewActionJumpToStart:
		logging.Infof("Jumping to start (if filtered will be first row in filter")
		m.jumpToStart()
	case viewActionJumpToEnd:
		logging.Infof("Jumping to end (if filtered will be last row in filter")
		m.jumpToEnd()
	case viewActionToggleShowMarks:
		// Show Marks only
		logging.Infof("Toggle for Show Marks Only has been pressed")
		m.table.showOnlyMarked = !m.table.showOnlyMarked
		m.recordChange("marks-only view")
		cmd = m.startFilterOperation(fmt.Sprintf("Show only marked: %t", m.table.showOnlyMarked))
	case viewActionNextMark:
		// Next mark jump
		logging.Debug("Here we go; jumping to the next mark")
		m.jumpToNextMark()
	case viewActionPrevMark:
		logging.Debug("Back once again: jumping to the previous mark")
		m.jumpToPreviousMark()
	case viewActionSearchNext:
		if !m.searchNext() {
			cmd = m.view.notice.Start("No matches", "warn", noticeDuration)
		}
		cmd = batchCmd(cmd, m.ensureSearchIndex())
	case viewActionSearchPrev:
		if !m.searchPrev() {
			cmd = m.view.notice.Start("No matches", "warn", noticeDuration)
		}
		cmd = batchCmd(cmd, m.ensureSearchIndex())
		m.ready = true
	case viewActionToggleFilter:
		logging.Infof("Shift F, toggling Filter")
		if !m.toggleFilterState() {
			cmd = m.view.notice.Start("No filter configured", "warn", noticeDuration)
			break
		}
		if m.table.filterEnabled {
			cmd = m.startFilterOperation("Filter enabled")
		} else {
			cmd = m.startFilterOperation("Filter disabled")
		}
	case viewActionToggleGraph:
		if m.graphConfig.Enabled {
			m.view.graphWindow.Open = !m.view.graphWindow.Open
			m.recordChange("graph view")
			m.refreshView("graph-toggle", true)
			didRefresh = true
		}
	case viewActionRowDown:
		if m.cursor < len(m.table.filteredIndices)-1 {
			m.cursor++
		}
	case viewActionRowUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case viewActionPageUp:
		m.pageUp()
	case viewActionPageDown:
		m.pageDown()
	case viewActionOpenPalette:
		return m, m.openCommandPalette()
	case viewActionOpenHelp:
		m.openHelpDialog()
		return m, nil
	case viewActionScrollLeft:
		m.scrollColumns(-4)
	case viewActionScrollRight:
		m.scrollColumns(4)
	case viewActionSave:
		m.openSaveDialog()
		return m, nil
	}

	//TODO: DON'T THINK WE SHOULD BE RENDERING TABLE EVERY TIME TBH
	if m.ready && !didRefresh {
		m.refreshView("view-key", false)
	}
	return m, cmd
}

func (m *Model) resolveViewAction(msg tea.KeyMsg) viewAction {
	switch {
	case key.Matches(msg, Keys.ColumnManager):
		return viewActionOpenColumnManager
	case key.Matches(msg, Keys.CommentOps):
		return viewActionPrefixComment
	case key.Matches(msg, Keys.TimeOps):
		return viewActionPrefixTime
	case key.Matches(msg, Keys.ExportOps):
		return viewActionPrefixExport
	case key.Matches(msg, Keys.JumpToLineNo):
		return viewActionJump
	case key.Matches(msg, Keys.Filter):
		return viewActionFilter
	case key.Matches(msg, Keys.Search):
		return viewActionSearch
	case key.Matches(msg, Keys.MarkMode):
		return viewActionMarkMode
	case key.Matches(msg, Keys.Quit):
		return viewActionQuit
	case m.quitConfirmationActive() && key.Matches(msg, Keys.ClearRange):
		return viewActionCancelQuit
	case key.Matches(msg, Keys.Undo):
		return viewActionUndo
	case key.Matches(msg, Keys.Redo):
		return viewActionRedo
	case key.Matches(msg, Keys.CopyRow):
		return viewActionCopyRow
	case key.Matches(msg, Keys.ToggleInspector):
		return viewActionToggleInspector
	case m.view.inspector.open && key.Matches(msg, Keys.InspectorNextField):
		return viewActionInspectorNextField
	case m.view.inspector.open && key.Matches(msg, Keys.InspectorPrevField):
		return viewActionInspectorPrevField
	case m.view.inspector.open && key.Matches(msg, Keys.InspectorCopyField):
		return viewActionInspectorCopyField
	case m.view.inspector.open && key.Matches(msg, Keys.InspectorScrollDown):
		return viewActionInspectorScrollDown
	case m.view.inspector.open && key.Matches(msg, Keys.InspectorScrollUp):
		return viewActionInspectorScrollUp
	case key.Matches(msg, Keys.SelectRange):
		return viewActionToggleRange
	case key.Matches(msg, Keys.ClearRange) && m.view.rowRange.active:
		return viewActionClearRange
	case key.Matches(msg, Keys.JumpToStart):
		return viewActionJumpToStart
	case key.Matches(msg, Keys.JumpToEnd):
		return viewActionJumpToEnd
	case key.Matches(msg, Keys.ShowMarksOnly):
		return viewActionToggleShowMarks
	case key.Matches(msg, Keys.NextMark):
		return viewActionNextMark
	case key.Matches(msg, Keys.PrevMark):
		return viewActionPrevMark
	case key.Matches(msg, Keys.SearchNext):
		return viewActionSearchNext
	case key.Matches(msg, Keys.SearchPrev):
		return viewActionSearchPrev
	case key.Matches(msg, Keys.ToggleFilter):
		return viewActionToggleFilter
	case key.Matches(msg, Keys.ToggleGraph):
		return viewActionToggleGraph
	case key.Matches(msg, Keys.RowDown):
		return viewActionRowDown
	case key.Matches(msg, Keys.RowUp):
		return viewActionRowUp
	case key.Matches(msg, Keys.PageUp):
		return viewActionPageUp
	case key.Matches(msg, Keys.PageDown):
		return viewActionPageDown
	case key.Matches(msg, Keys.CommandPalette):
		return viewActionOpenPalette
	case key.Matches(msg, Keys.OpenHelp):
		return viewActionOpenHelp
	case key.Matches(msg, Keys.ScrollLeft):
		return viewActionScrollLeft
	case key.Matches(msg, Keys.ScrollRight):
		return viewActionScrollRight
	case key.Matches(msg, Keys.SaveToFile):
		return viewActionSave
	default:
		return viewActionNone
	}
}

func (m *Model) confirmOrQuit() tea.Cmd {
	if !m.dirty {
		if err := m.removeRecoverySnapshotNow(); err != nil {
			logging.Errorf("remove clean recovery snapshot before quit: %v", err)
		}
		return tea.Quit
	}
	now := time.Now()
	if now.After(m.view.quitConfirmUntil) {
		m.view.quitConfirmUntil = now.Add(quitConfirmWindow)
		return m.view.notice.Start("Unsaved changes: s save, q quit without saving, esc cancel", "warn", quitConfirmWindow)
	}
	if err := m.writeRecoverySnapshotNow(); err != nil {
		logging.Errorf("final recovery snapshot before quit: %v", err)
	}
	return tea.Quit
}

func (m *Model) quitConfirmationActive() bool {
	return !m.view.quitConfirmUntil.IsZero() && time.Now().Before(m.view.quitConfirmUntil)
}

func (m *Model) handleViewPrefixKey(msg tea.KeyMsg) (handled bool, cmd tea.Cmd, refresh bool) {
	if m.view.pendingViewPrefix == "" {
		return false, nil, false
	}
	defer func() {
		m.clearPrefixHint()
		m.view.pendingViewPrefix = ""
	}()

	switch m.resolveViewPrefixAction(msg) {
	case viewPrefixActionCommentEdit:
		return true, m.enterCommand(CmdComment, "", true, false), false
	case viewPrefixActionCommentToggleDrawer:
		m.view.drawerOpen = !m.view.drawerOpen
		if m.view.drawerOpen {
			m.view.inspector.open = false
		}
		logging.Infof("handleViewPrefixKey: toggled comment drawer to %t", m.view.drawerOpen)
		return true, nil, true
	case viewPrefixActionTimeWindowOpen:
		m.openTimeWindowDrawer()
		return true, nil, false
	case viewPrefixActionTimeSetStart:
		return true, m.quickSetTimeWindowEdge(true), true
	case viewPrefixActionTimeSetEnd:
		return true, m.quickSetTimeWindowEdge(false), true
	case viewPrefixActionTimeReset:
		return true, m.quickResetTimeWindow(), true
	case viewPrefixActionExportData:
		m.openExportDialog()
		return true, nil, false
	case viewPrefixActionExportGraph:
		if !m.graphConfig.Enabled {
			return true, m.view.notice.Start("Graph export is not available", "warn", noticeDuration), false
		}
		return true, m.startGraphExportOperation(defaultGraphExportPath(*m)), false
	case viewPrefixActionCancel:
		return true, nil, false
	}
	return false, nil, false
}

func (m *Model) resolveViewPrefixAction(msg tea.KeyMsg) viewPrefixAction {
	keyStr := strings.ToLower(strings.TrimSpace(msg.String()))
	switch {
	case keyStr == "esc":
		return viewPrefixActionCancel
	case m.view.pendingViewPrefix == "c" && keyStr == "e":
		return viewPrefixActionCommentEdit
	case m.view.pendingViewPrefix == "c" && keyStr == "v":
		return viewPrefixActionCommentToggleDrawer
	case m.view.pendingViewPrefix == "t" && keyStr == "w":
		return viewPrefixActionTimeWindowOpen
	case m.view.pendingViewPrefix == "t" && keyStr == "b":
		return viewPrefixActionTimeSetStart
	case m.view.pendingViewPrefix == "t" && keyStr == "e":
		return viewPrefixActionTimeSetEnd
	case m.view.pendingViewPrefix == "t" && keyStr == "r":
		return viewPrefixActionTimeReset
	case m.view.pendingViewPrefix == "e" && keyStr == "d":
		return viewPrefixActionExportData
	case m.view.pendingViewPrefix == "e" && keyStr == "g":
		return viewPrefixActionExportGraph
	default:
		return viewPrefixActionNone
	}
}

func (m *Model) quickSetTimeWindowEdge(setStart bool) tea.Cmd {
	if !m.table.hasTimeBounds {
		return m.view.notice.Start("No timestamps available", "warn", noticeDuration)
	}
	ts, ok := m.cursorTimestamp()
	if !ok {
		return m.view.notice.Start("No timestamp on current row", "warn", noticeDuration)
	}
	m.setTimeWindowEdge(ts, setStart)
	label := "end"
	if setStart {
		label = "start"
	}
	return m.view.notice.Start(fmt.Sprintf("Window %s set", label), "", noticeDuration)
}

func (m *Model) quickResetTimeWindow() tea.Cmd {
	if !m.table.hasTimeBounds {
		return m.view.notice.Start("No timestamps available", "warn", noticeDuration)
	}
	previous := m.table.timeWindow
	m.table.timeWindow.Enabled = true
	m.table.timeWindow.Start = m.table.timeMin
	m.table.timeWindow.End = m.table.timeMax
	if previous != m.table.timeWindow {
		m.recordChange("time window")
	}
	m.view.timeWindow.DraftStart = m.table.timeWindow.Start
	m.view.timeWindow.DraftEnd = m.table.timeWindow.End
	m.updateTimeWindowInputsFromDraft()
	m.applyFilter()
	return m.view.notice.Start("Window reset", "", noticeDuration)
}
