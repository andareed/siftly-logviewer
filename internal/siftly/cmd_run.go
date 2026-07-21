package siftly

import (
	"strconv"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

type commandRunner func(*Model, string) tea.Cmd

var commandRunnerTable = map[Command]commandRunner{
	CmdJump:        runJumpCommand,
	CmdSearch:      runSearchCommand,
	CmdFilter:      runFilterCommand,
	CmdSort:        runSortCommand,
	CmdColumns:     runColumnsCommand,
	CmdColumnOrder: runColumnOrderCommand,
	CmdComment:     runCommentCommand,
}

func (m *Model) runCommand() tea.Cmd {
	runner, ok := commandRunnerTable[m.view.command.cmd]
	if !ok {
		return nil
	}
	return runner(m, m.commandValue())
}

func runJumpCommand(m *Model, raw string) tea.Cmd {
	if n, err := strconv.Atoi(raw); err == nil {
		return m.jumpToLine(n)
	}
	return m.view.notice.Start("Invalid line number", "warn", noticeDuration)
}

func runSearchCommand(m *Model, raw string) tea.Cmd {
	indexCmd := m.setSearchQuery(raw)
	if m.searchNext() {
		return indexCmd
	}
	return batchCmd(indexCmd, m.view.notice.Start("No matches", "warn", noticeDuration))
}

func runFilterCommand(m *Model, raw string) tea.Cmd {
	if err := m.prepareFilterPattern(raw); err != nil {
		return m.view.notice.Start("Invalid filter", "warn", noticeDuration)
	}
	if err := m.recordFilterHistory(raw); err != nil {
		return m.startFilterOperationWithKind("Filter applied; history error", "warn")
	}
	return m.startFilterOperation("Filter applied")
}

func runSortCommand(m *Model, raw string) tea.Cmd {
	if err := m.setSortSpec(raw); err != nil {
		return m.view.notice.Start(err.Error(), "warn", noticeDuration)
	}
	return nil
}

func runColumnsCommand(m *Model, raw string) tea.Cmd {
	buf := strings.TrimSpace(raw)
	if buf == "" {
		return m.view.notice.Start("Enter column names or numbers", "warn", noticeDuration)
	}
	m.view.lastColumnsSpec = buf
	if strings.EqualFold(buf, "all") {
		changed := false
		for i := range m.table.header {
			changed = changed || !m.table.header[i].Visible
			m.table.header[i].Visible = true
		}
		if changed {
			m.recordChange("column visibility")
		}
		m.refreshView("show-all-columns", true)
		return m.view.notice.Start("All columns shown", "", noticeDuration)
	}

	toggled, missing, err := m.toggleColumnsBySpec(buf)
	if err != nil {
		return m.view.notice.Start(err.Error(), "warn", noticeDuration)
	}
	if len(toggled) > 0 {
		m.recordChange("column visibility")
	}
	return m.view.notice.Start(columnsNoticeText(toggled, missing), "", noticeDuration)
}

func runColumnOrderCommand(m *Model, raw string) tea.Cmd {
	buf := strings.TrimSpace(raw)
	if buf == "" {
		return m.view.notice.Start("Enter a column order", "warn", noticeDuration)
	}
	ordered, missing, err := m.reorderColumnsBySpec(buf)
	if err != nil {
		return m.view.notice.Start(err.Error(), "warn", noticeDuration)
	}
	if len(ordered) == 0 {
		return m.view.notice.Start("No columns reordered", "warn", noticeDuration)
	}
	m.recordChange("column order")
	if len(missing) > 0 {
		return m.view.notice.Start("Reordered columns; missing: "+strings.Join(missing, ", "), "warn", noticeDuration)
	}
	return m.view.notice.Start("Reordered columns", "", noticeDuration)
}

func runCommentCommand(m *Model, raw string) tea.Cmd {
	m.addComment(raw)
	return m.view.notice.Start("Comment added", "", noticeDuration)
}

func (m *Model) handleCommandInputMsg(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if m.view.mode != modeCommand || !commandUsesTextInput(m.view.command.cmd) {
		return m, nil, false
	}
	return m, m.updateCommandInput(msg), true
}

func (m *Model) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.view.command.cmd == CmdFilter {
		switch msg.String() {
		case "ctrl+p":
			return m, m.openFilterPalette(false)
		case "ctrl+h":
			return m, m.openFilterPalette(true)
		}
	}

	// universal cancel
	if msg.Type == tea.KeyEsc {
		cmd := m.exitCommand(true)
		return m, cmd
	}

	if m.view.command.cmd == CmdMark {
		return m.handleMarkCommandKey(msg)
	}
	if m.view.command.cmd == CmdTimeWindowSet {
		return m.handleTimeWindowSetCommandKey(msg)
	}

	// commit
	if msg.Type == tea.KeyEnter {
		cmd := m.runCommand() // returns tea.Cmd or nil
		exitCmd := m.exitCommand(true)
		return m, tea.Batch(cmd, exitCmd)
	}

	if !commandUsesTextInput(m.view.command.cmd) {
		return m, nil
	}
	return m, m.updateCommandInput(msg)
}

func (m *Model) updateCommandInput(msg tea.Msg) tea.Cmd {
	m.ensureCommandTextInput()
	var cmd tea.Cmd
	m.view.command.input, cmd = m.view.command.input.Update(msg)

	// The footer renders textinput.Value(), not textinput.View(), so cursor
	// blink commands only create needless internal messages. Preserve ctrl+v
	// because textinput uses a command to fetch clipboard contents.
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+v" {
		return cmd
	}
	return nil
}

func (m *Model) openFilterPalette(history bool) tea.Cmd {
	cfg, err := m.loadFilterConfig()
	if err != nil {
		return m.view.notice.Start("Filter palette error", "warn", noticeDuration)
	}
	presets := make([]dialogs.FilterPreset, 0, len(cfg.Presets))
	for _, preset := range cfg.Presets {
		presets = append(presets, dialogs.FilterPreset{
			Pattern:     preset.Pattern,
			Description: preset.Description,
		})
	}

	if history {
		m.activeDialog = dialogs.NewFilterHistoryPaletteDialog(
			presets,
			cfg.History,
			m.terminalWidth,
			m.terminalHeight,
			m.styles.RowSelectedFG,
			m.styles.RowSelectedBG,
		)
	} else {
		m.activeDialog = dialogs.NewFilterPaletteDialog(
			presets,
			cfg.History,
			m.terminalWidth,
			m.terminalHeight,
			m.styles.RowSelectedFG,
			m.styles.RowSelectedBG,
		)
	}
	m.activeDialog.Show()
	return nil
}
