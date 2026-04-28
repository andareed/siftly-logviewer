package siftly

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Command int

const (
	CmdNone Command = iota
	CmdJump
	CmdSearch
	CmdFilter
	CmdSort
	CmdComment
	CmdMark
	CmdTimeWindowSet
	CmdColumns
	CmdColumnOrder
)

type CommandInput struct {
	cmd Command
	buf string
}

type commandMeta struct {
	Label  string
	Badge  string
	Prompt string
	Hints  string
}

var commandMetaTable = map[Command]commandMeta{
	CmdNone: {
		Label:  "NORMAL",
		Badge:  "[-]",
		Prompt: "",
		Hints:  "enter: apply   esc: cancel",
	},
	CmdJump: {
		Label:  "JUMP",
		Badge:  "[:]",
		Prompt: "line: ",
		Hints:  "enter: jump   esc: cancel",
	},
	CmdSearch: {
		Label:  "SEARCH",
		Badge:  "[/]",
		Prompt: "search: ",
		Hints:  "enter: next match   esc: cancel",
	},
	CmdFilter: {
		Label:  "FILTER",
		Badge:  "[f]",
		Prompt: "filter: ",
		Hints:  "enter: apply   ctrl+p: palette   esc: cancel",
	},
	CmdSort: {
		Label:  "SORT",
		Badge:  "[S]",
		Prompt: "sort: ",
		Hints:  "enter: apply   off/empty: reset   esc: cancel",
	},
	CmdComment: {
		Label:  "COMMENT",
		Badge:  "[c]",
		Prompt: "comment: ",
		Hints:  "enter: save comment   esc: cancel",
	},
	CmdMark: {
		Label:  "MARK",
		Badge:  "[m]",
		Prompt: "mark: ",
		Hints:  "r/g/a: mark   c: clear   esc: cancel",
	},
	CmdTimeWindowSet: {
		Label:  "WINDOW",
		Badge:  "[T]",
		Prompt: "window: ",
		Hints:  "b: set start   e: set end   r: reset   esc: cancel",
	},
	CmdColumns: {
		Label:  "COLUMNS",
		Badge:  "[C]",
		Prompt: "cols: ",
		Hints:  "enter: toggle columns   all: show all   esc: cancel",
	},
	CmdColumnOrder: {
		Label:  "ORDER",
		Badge:  "[O]",
		Prompt: "order: ",
		Hints:  "enter: apply order   esc: cancel",
	},
}

func lookupCommandMeta(cmd Command) commandMeta {
	meta, ok := commandMetaTable[cmd]
	if ok {
		return meta
	}
	return commandMetaTable[CmdNone]
}

func commandLabel(cmd Command) string { return lookupCommandMeta(cmd).Label }

func (m *Model) commandBadge(cmd Command) string { return lookupCommandMeta(cmd).Badge }

func (m *Model) commandPrompt(cmd Command) string { return lookupCommandMeta(cmd).Prompt }

func (m *Model) commandHintsLine(cmd Command) string { return lookupCommandMeta(cmd).Hints }

func (m *Model) commandSeed(cmd Command) string {
	switch cmd {
	case CmdFilter:
		if m.table.filterPattern != "" {
			return m.table.filterPattern
		}
		return ""
	case CmdSort:
		return m.currentSortSeed()
	case CmdSearch:
		return m.view.searchQuery
	case CmdComment:
		return m.getCommentContent(m.currentRowHashID())
	case CmdColumnOrder:
		return m.currentColumnOrderSeed()
	default:
		return ""
	}
}

func commandUsesMainBodySnapshot(cmd Command) bool {
	switch cmd {
	case CmdJump, CmdSearch, CmdFilter, CmdSort, CmdComment, CmdColumns, CmdColumnOrder:
		return true
	default:
		return false
	}
}

// activeCommandLine returns the command prompt text for the footer status line.
func (m *Model) activeCommandLine() string {
	badge := m.commandBadge(m.view.command.cmd)
	prompt := m.commandPrompt(m.view.command.cmd)
	return badge + " " + prompt + m.view.command.buf + m.commandPreviewSuffix()
}

// enterCommand switches the UI to command mode, seeds the input buffer,
// and optionally refreshes the view or shows a hint notice.
func (m *Model) enterCommand(cmd Command, seed string, showHint bool, refresh bool) tea.Cmd {
	m.view.command.cmd = cmd
	if seed != "" {
		m.view.command.buf = seed
	} else {
		m.view.command.buf = m.commandSeed(cmd)
	}

	m.view.mode = modeCommand
	if refresh {
		m.refreshView("enter-command", false)
	}
	if commandUsesMainBodySnapshot(cmd) {
		m.captureMainBodySnapshot(m.panelWidth())
	}
	if showHint {
		m.setModeHint(m.commandHintsLine(cmd))
		return nil
	}
	return nil
}

func (m *Model) exitCommand(refresh bool) tea.Cmd {
	m.clearModeHint()
	m.clearMainBodySnapshot()
	m.view.command = CommandInput{}
	m.view.mode = modeView
	if refresh {
		m.refreshView("exit-command", false)
	}
	return nil
}
