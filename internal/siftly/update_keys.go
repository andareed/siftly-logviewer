package siftly

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type viewAction int

const (
	viewActionNone viewAction = iota
	viewActionPrefixView
	viewActionPrefixComment
	viewActionPrefixTime
	viewActionJump
	viewActionFilter
	viewActionSearch
	viewActionMarkMode
	viewActionQuit
	viewActionCopyRow
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
	viewActionOpenHelp
	viewActionScrollLeft
	viewActionScrollRight
	viewActionSave
	viewActionExport
)

type viewPrefixAction int

const (
	viewPrefixActionNone viewPrefixAction = iota
	viewPrefixActionColumns
	viewPrefixActionSort
	viewPrefixActionColumnOrder
	viewPrefixActionResetLayout
	viewPrefixActionCommentEdit
	viewPrefixActionCommentToggleDrawer
	viewPrefixActionTimeWindowOpen
	viewPrefixActionTimeSetStart
	viewPrefixActionTimeSetEnd
	viewPrefixActionTimeReset
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

	switch m.resolveViewAction(msg) {
	case viewActionPrefixView:
		m.view.pendingViewPrefix = "v"
		m.setPrefixHint("c: columns   s: sort   o: order   r: reset   esc: cancel")
		cmd = nil
		return m, cmd
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
		return m, tea.Quit
	case viewActionCopyRow:
		logging.Infof("Key Combination for CopyRow To Clipboard")
		cmd = m.copyRowToClipboard()
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
		cmd = m.view.notice.Start(fmt.Sprintf("'Show Only Marked Rows' toggled {%t}", m.table.showOnlyMarked), "", noticeDuration)
		m.applyFilter()
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
	case viewActionSearchPrev:
		if !m.searchPrev() {
			cmd = m.view.notice.Start("No matches", "warn", noticeDuration)
		}
		m.ready = true
	case viewActionToggleFilter:
		logging.Infof("Shift F, toggling Filter")
		if !m.toggleFilter() {
			cmd = m.view.notice.Start("No filter configured", "warn", noticeDuration)
			break
		}
		if m.table.filterEnabled {
			cmd = m.view.notice.Start("Filter enabled", "", noticeDuration)
		} else {
			cmd = m.view.notice.Start("Filter disabled", "", noticeDuration)
		}
	case viewActionToggleGraph:
		if m.graphConfig.Enabled {
			m.view.graphWindow.Open = !m.view.graphWindow.Open
			m.refreshView("graph-toggle", true)
			didRefresh = true
		}
	case viewActionRowDown:
		if m.cursor < len(m.table.rows)-1 {
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
	case viewActionOpenHelp:
		m.openHelpDialog()
		return m, nil
	case viewActionScrollLeft:
		m.viewport.ScrollLeft(4) // tune step
	case viewActionScrollRight:
		m.viewport.ScrollRight(4)
	case viewActionSave:
		m.openSaveDialog()
		return m, nil
	case viewActionExport:
		m.openExportDialog()
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
	case key.Matches(msg, Keys.ColumnViewOps):
		return viewActionPrefixView
	case key.Matches(msg, Keys.CommentOps):
		return viewActionPrefixComment
	case key.Matches(msg, Keys.TimeOps):
		return viewActionPrefixTime
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
	case key.Matches(msg, Keys.CopyRow):
		return viewActionCopyRow
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
	case key.Matches(msg, Keys.OpenHelp):
		return viewActionOpenHelp
	case key.Matches(msg, Keys.ScrollLeft):
		return viewActionScrollLeft
	case key.Matches(msg, Keys.ScrollRight):
		return viewActionScrollRight
	case key.Matches(msg, Keys.SaveToFile):
		return viewActionSave
	case key.Matches(msg, Keys.ExportToFile):
		return viewActionExport
	default:
		return viewActionNone
	}
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
	case viewPrefixActionColumns:
		seed := m.view.lastColumnsSpec
		return true, m.enterCommand(CmdColumns, seed, true, false), false
	case viewPrefixActionSort:
		return true, m.enterCommand(CmdSort, "", true, false), false
	case viewPrefixActionColumnOrder:
		return true, m.enterCommand(CmdColumnOrder, m.currentColumnOrderSeed(), true, false), false
	case viewPrefixActionResetLayout:
		return true, m.resetViewLayoutState(), true
	case viewPrefixActionCommentEdit:
		return true, m.enterCommand(CmdComment, "", true, false), false
	case viewPrefixActionCommentToggleDrawer:
		m.view.drawerOpen = !m.view.drawerOpen
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
	case m.view.pendingViewPrefix == "v" && keyStr == "c":
		return viewPrefixActionColumns
	case m.view.pendingViewPrefix == "v" && keyStr == "s":
		return viewPrefixActionSort
	case m.view.pendingViewPrefix == "v" && keyStr == "o":
		return viewPrefixActionColumnOrder
	case m.view.pendingViewPrefix == "v" && keyStr == "r":
		return viewPrefixActionResetLayout
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
	m.table.timeWindow.Enabled = true
	m.table.timeWindow.Start = m.table.timeMin
	m.table.timeWindow.End = m.table.timeMax
	m.view.timeWindow.DraftStart = m.table.timeWindow.Start
	m.view.timeWindow.DraftEnd = m.table.timeWindow.End
	m.updateTimeWindowInputsFromDraft()
	m.applyFilter()
	return m.view.notice.Start("Window reset", "", noticeDuration)
}
