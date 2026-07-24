package siftly

import (
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

type paletteCommandSpec struct {
	item     dialogs.CommandItem
	sequence []tea.KeyMsg
	execute  func() tea.Cmd
}

const (
	commandPaletteOpen         = "general.palette"
	commandHelpOpen            = "general.help"
	commandQuit                = "general.quit"
	commandUndo                = "history.undo"
	commandRedo                = "history.redo"
	commandRowUp               = "navigation.row-up"
	commandRowDown             = "navigation.row-down"
	commandPageUp              = "navigation.page-up"
	commandPageDown            = "navigation.page-down"
	commandScrollLeft          = "navigation.scroll-left"
	commandScrollRight         = "navigation.scroll-right"
	commandJumpStart           = "navigation.start"
	commandJumpEnd             = "navigation.end"
	commandJumpLine            = "navigation.line"
	commandSearch              = "search.open"
	commandSearchNext          = "search.next"
	commandSearchPrevious      = "search.previous"
	commandFilter              = "filter.open"
	commandFilterToggle        = "filter.toggle"
	commandFilterPresets       = "filter.presets"
	commandFilterHistory       = "filter.history"
	commandRangeToggle         = "rows.range-toggle"
	commandRangeClear          = "rows.range-clear"
	commandCopyRows            = "rows.copy"
	commandMarkOpen            = "marks.open"
	commandMarkRed             = "marks.red"
	commandMarkGreen           = "marks.green"
	commandMarkAmber           = "marks.amber"
	commandMarkClear           = "marks.clear"
	commandMarksOnly           = "marks.only"
	commandMarkNext            = "marks.next"
	commandMarkPrevious        = "marks.previous"
	commandCommentEdit         = "comments.edit"
	commandCommentDrawer       = "comments.drawer"
	commandInspectorToggle     = "inspector.toggle"
	commandInspectorNextField  = "inspector.next-field"
	commandInspectorPrevField  = "inspector.previous-field"
	commandInspectorCopyField  = "inspector.copy-field"
	commandInspectorScrollDown = "inspector.scroll-down"
	commandInspectorScrollUp   = "inspector.scroll-up"
	commandColumns             = "view.columns"
	commandTimeWindow          = "time.window"
	commandTimeStart           = "time.start"
	commandTimeEnd             = "time.end"
	commandTimeReset           = "time.reset"
	commandGraphToggle         = "graph.toggle"
	commandSave                = "output.save"
	commandExportData          = "output.data"
	commandExportGraph         = "output.graph"
	commandReloadFullSource    = "data.reload-full"
)

func (m *Model) commandCatalog() []paletteCommandSpec {
	commands := make([]paletteCommandSpec, 0, 48)
	add := func(id, category, title, shortcut, description, keywords string, sequence ...tea.KeyMsg) {
		commands = append(commands, paletteCommandSpec{
			item: dialogs.CommandItem{
				ID:          id,
				Category:    category,
				Title:       title,
				Shortcut:    shortcut,
				Description: description,
				Keywords:    keywords,
				Enabled:     true,
			},
			sequence: sequence,
		})
	}
	addDirect := func(id, category, title, shortcut, description, keywords string, execute func() tea.Cmd) {
		commands = append(commands, paletteCommandSpec{
			item: dialogs.CommandItem{
				ID:          id,
				Category:    category,
				Title:       title,
				Shortcut:    shortcut,
				Description: description,
				Keywords:    keywords,
				Enabled:     true,
			},
			execute: execute,
		})
	}
	disable := func(id, reason string) {
		for i := range commands {
			if commands[i].item.ID == id {
				commands[i].item.Enabled = false
				commands[i].item.DisabledReason = reason
				return
			}
		}
	}

	add(commandPaletteOpen, "General", "Open command palette", "p", "Search and run any available command", "commands actions", runeKey('p'))
	add(commandHelpOpen, "General", "Open keyboard reference", "?", "Browse every command and shortcut by category", "keys help shortcuts reference", runeKey('?'))
	add(commandQuit, "General", "Quit", "q", "Exit Siftly, confirming when changes are unsaved", "close exit", runeKey('q'))
	add(commandUndo, "History", "Undo last change", "u", "Restore the previous annotation or view state", "history revert", runeKey('u'))
	add(commandRedo, "History", "Redo last change", "r", "Reapply the most recently undone change", "history restore", runeKey('r'))
	if !m.canUndo() {
		disable(commandUndo, "Nothing to undo")
	}
	if !m.canRedo() {
		disable(commandRedo, "Nothing to redo")
	}

	add(commandRowUp, "Navigation", "Move up one row", "k / Up", "Move the cursor to the previous displayed row", "previous", runeKey('k'))
	add(commandRowDown, "Navigation", "Move down one row", "j / Down", "Move the cursor to the next displayed row", "next", runeKey('j'))
	add(commandPageUp, "Navigation", "Page up", "Ctrl+U / PgUp", "Move up by one page", "previous page", keyType(tea.KeyCtrlU))
	add(commandPageDown, "Navigation", "Page down", "Ctrl+D / PgDn", "Move down by one page", "next page", keyType(tea.KeyCtrlD))
	add(commandScrollLeft, "Navigation", "Scroll columns left", "h / Left", "Move the table viewport left", "horizontal", runeKey('h'))
	add(commandScrollRight, "Navigation", "Scroll columns right", "l / Right", "Move the table viewport right", "horizontal", runeKey('l'))
	add(commandJumpStart, "Navigation", "Jump to first displayed row", "g / Home", "Move to the beginning of the current result set", "top first", runeKey('g'))
	add(commandJumpEnd, "Navigation", "Jump to last displayed row", "G / End", "Move to the end of the current result set", "bottom last", runeKey('G'))
	add(commandJumpLine, "Navigation", "Jump to source line", ":", "Enter an original source line number", "goto row number", runeKey(':'))
	if len(m.table.filteredIndices) == 0 {
		for _, id := range []string{commandRowUp, commandRowDown, commandPageUp, commandPageDown, commandJumpStart, commandJumpEnd, commandJumpLine} {
			disable(id, "No displayed rows")
		}
	}

	add(commandSearch, "Search and Filter", "Search displayed rows", "/", "Find text without changing the displayed rows", "find", runeKey('/'))
	add(commandSearchNext, "Search and Filter", "Next search match", "n", "Move to the next search result", "find forward", runeKey('n'))
	add(commandSearchPrevious, "Search and Filter", "Previous search match", "N", "Move to the previous search result", "find backward", runeKey('N'))
	add(commandFilter, "Search and Filter", "Filter rows by regular expression", "f", "Restrict displayed rows using whole-row matching", "regex", runeKey('f'))
	add(commandFilterToggle, "Search and Filter", "Toggle current filter", "F", "Enable or disable the configured filter", "show hide", runeKey('F'))
	add(commandFilterPresets, "Search and Filter", "Open filter presets", "f Ctrl+P", "Choose or search configured filter presets", "saved filters", runeKey('f'), keyType(tea.KeyCtrlP))
	add(commandFilterHistory, "Search and Filter", "Open filter history", "f Ctrl+R / Up", "Choose or search previously used filters", "recent filters", runeKey('f'), keyType(tea.KeyCtrlR))
	if strings.TrimSpace(m.view.searchQuery) == "" {
		disable(commandSearchNext, "No active search")
		disable(commandSearchPrevious, "No active search")
	}
	if strings.TrimSpace(m.table.filterPattern) == "" {
		disable(commandFilterToggle, "No filter configured")
	}

	add(commandRangeToggle, "Rows and Marks", "Start or clear range selection", "Space", "Anchor a displayed-row selection at the current row", "multi select range", runeKey(' '))
	add(commandRangeClear, "Rows and Marks", "Clear range selection", "Esc", "Remove the active displayed-row selection", "deselect", keyType(tea.KeyEsc))
	add(commandCopyRows, "Rows and Marks", "Copy current row or selection", "Ctrl+C", "Copy the current row or selected rows", "clipboard yank", keyType(tea.KeyCtrlC))
	add(commandMarkOpen, "Rows and Marks", "Open mark command", "m [count] r/g/a/c", "Mark selected rows or the current and next count displayed rows", "colour color count", runeKey('m'))
	add(commandMarkRed, "Rows and Marks", "Mark rows red", "m r", "Apply a red mark to the current row or selection", "colour color", runeKey('m'), runeKey('r'))
	add(commandMarkGreen, "Rows and Marks", "Mark rows green", "m g", "Apply a green mark to the current row or selection", "colour color", runeKey('m'), runeKey('g'))
	add(commandMarkAmber, "Rows and Marks", "Mark rows amber", "m a", "Apply an amber mark to the current row or selection", "colour color yellow", runeKey('m'), runeKey('a'))
	add(commandMarkClear, "Rows and Marks", "Clear row marks", "m c", "Remove marks from the current row or selection", "unmark", runeKey('m'), runeKey('c'))
	add(commandMarksOnly, "Rows and Marks", "Toggle marked rows only", "M", "Show all rows or only rows carrying a mark", "filter marks", runeKey('M'))
	add(commandMarkNext, "Rows and Marks", "Jump to next marked row", "]", "Move to the next displayed marked row", "mark forward", runeKey(']'))
	add(commandMarkPrevious, "Rows and Marks", "Jump to previous marked row", "[", "Move to the previous displayed marked row", "mark backward", runeKey('['))
	if !m.view.rowRange.active {
		disable(commandRangeClear, "No active row selection")
	}
	if len(m.table.filteredIndices) == 0 {
		for _, id := range []string{commandRangeToggle, commandCopyRows, commandMarkOpen, commandMarkRed, commandMarkGreen, commandMarkAmber, commandMarkClear} {
			disable(id, "No displayed rows")
		}
	}
	if len(m.table.markedRows) == 0 {
		disable(commandMarkNext, "No marked rows")
		disable(commandMarkPrevious, "No marked rows")
	}

	add(commandCommentEdit, "Comments and Inspector", "Edit row comment", "c e", "Add, change, or clear the current row comment", "annotation note", runeKey('c'), runeKey('e'))
	add(commandCommentDrawer, "Comments and Inspector", "Toggle comment drawer", "c v", "Show or hide comments alongside the table", "annotation notes", runeKey('c'), runeKey('v'))
	add(commandInspectorToggle, "Comments and Inspector", "Toggle row inspector", "Enter", "Show or hide complete details for the current row", "details fields", keyType(tea.KeyEnter))
	add(commandInspectorNextField, "Comments and Inspector", "Select next inspector field", "Tab", "Move inspector focus to the next field", "details", keyType(tea.KeyTab))
	add(commandInspectorPrevField, "Comments and Inspector", "Select previous inspector field", "Shift+Tab", "Move inspector focus to the previous field", "details", keyType(tea.KeyShiftTab))
	add(commandInspectorCopyField, "Comments and Inspector", "Copy inspector field", "y", "Copy the focused field value", "details clipboard", runeKey('y'))
	add(commandInspectorScrollDown, "Comments and Inspector", "Scroll inspector down", "J", "Scroll inspector content down", "details", runeKey('J'))
	add(commandInspectorScrollUp, "Comments and Inspector", "Scroll inspector up", "K", "Scroll inspector content up", "details", runeKey('K'))
	if len(m.table.filteredIndices) == 0 {
		disable(commandCommentEdit, "No displayed rows")
		disable(commandInspectorToggle, "No displayed rows")
	}
	if !m.view.inspector.open {
		for _, id := range []string{commandInspectorNextField, commandInspectorPrevField, commandInspectorCopyField, commandInspectorScrollDown, commandInspectorScrollUp} {
			disable(id, "Row inspector is closed")
		}
	}

	add(commandColumns, "View and Columns", "Manage columns", "v", "Search, show, order, sort, freeze, auto-fit, and reset columns", "layout fields ascending descending pin width reset defaults", runeKey('v'))

	add(commandTimeWindow, "Time Window", "Open time window", "t w", "Inspect and adjust the displayed time range", "date range", runeKey('t'), runeKey('w'))
	add(commandTimeStart, "Time Window", "Set window start from cursor", "t b", "Use the current row timestamp as the window start", "begin date range", runeKey('t'), runeKey('b'))
	add(commandTimeEnd, "Time Window", "Set window end from cursor", "t e", "Use the current row timestamp as the window end", "date range", runeKey('t'), runeKey('e'))
	add(commandTimeReset, "Time Window", "Reset time window", "t r", "Restore the complete source time range", "date range clear", runeKey('t'), runeKey('r'))
	if !m.table.hasTimeBounds {
		for _, id := range []string{commandTimeWindow, commandTimeStart, commandTimeEnd, commandTimeReset} {
			disable(id, "No timestamp range is available")
		}
	}

	if m.graphConfig.Enabled {
		add(commandGraphToggle, "Graph", "Toggle graph", "w", "Show or hide the configured graph", "chart plot", runeKey('w'))
	}

	add(commandSave, "Output and Data", "Save Siftly JSON", "s", "Write a reloadable Siftly file with data and annotations", "snapshot session", runeKey('s'))
	add(commandExportData, "Output and Data", "Export filtered data", "e d", "Write the currently filtered rows to CSV", "csv rows", runeKey('e'), runeKey('d'))
	if m.graphConfig.Enabled {
		add(commandExportGraph, "Output and Data", "Export graph SVG", "e g", "Write the current graph to a timestamped SVG", "chart plot image", runeKey('e'), runeKey('g'))
	}
	if m.CanReloadFullSource() {
		addDirect(commandReloadFullSource, "Output and Data", "Reload full source data", "Palette only", "Replace prefiltered data with the complete source", "refresh load", m.startFullSourceReloadOperation)
	}

	return commands
}

func (m *Model) commandItems() []dialogs.CommandItem {
	catalog := m.commandCatalog()
	items := make([]dialogs.CommandItem, len(catalog))
	for i := range catalog {
		items[i] = catalog[i].item
	}
	return items
}

func (m *Model) runPaletteCommand(id string) tea.Cmd {
	for _, command := range m.commandCatalog() {
		if command.item.ID != id {
			continue
		}
		if !command.item.Enabled {
			message := command.item.DisabledReason
			if strings.TrimSpace(message) == "" {
				message = "Command is unavailable"
			}
			return m.view.notice.Start(message, "warn", noticeDuration)
		}
		if command.execute != nil {
			return command.execute()
		}
		commands := make([]tea.Cmd, 0, len(command.sequence))
		for _, keyMsg := range command.sequence {
			_, cmd := m.handleKeyMsg(keyMsg)
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}
		switch len(commands) {
		case 0:
			return nil
		case 1:
			return commands[0]
		default:
			return tea.Sequence(commands...)
		}
	}
	return m.view.notice.Start("Command is no longer available", "warn", noticeDuration)
}

func runeKey(value rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{value}}
}

func keyType(value tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: value}
}
