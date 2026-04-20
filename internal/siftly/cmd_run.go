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
	return runner(m, m.view.command.buf)
}

func runJumpCommand(m *Model, raw string) tea.Cmd {
	if n, err := strconv.Atoi(raw); err == nil {
		return m.jumpToLine(n)
	}
	return m.view.notice.Start("Invalid line number", "warn", noticeDuration)
}

func runSearchCommand(m *Model, raw string) tea.Cmd {
	m.setSearchQuery(raw)
	if m.searchNext() {
		return nil
	}
	return m.view.notice.Start("No matches", "warn", noticeDuration)
}

func runFilterCommand(m *Model, raw string) tea.Cmd {
	if err := m.setFilterPattern(raw); err != nil {
		return m.view.notice.Start("Invalid filter", "warn", noticeDuration)
	}
	if err := m.recordFilterHistory(raw); err != nil {
		return m.view.notice.Start("Filter history error", "warn", noticeDuration)
	}
	return nil
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
		for i := range m.table.header {
			m.table.header[i].Visible = true
		}
		m.refreshView("show-all-columns", true)
		return m.view.notice.Start("All columns shown", "", noticeDuration)
	}

	toggled, missing, err := m.toggleColumnsBySpec(buf)
	if err != nil {
		return m.view.notice.Start(err.Error(), "warn", noticeDuration)
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
	if len(missing) > 0 {
		return m.view.notice.Start("Reordered columns; missing: "+strings.Join(missing, ", "), "warn", noticeDuration)
	}
	return m.view.notice.Start("Reordered columns", "", noticeDuration)
}

func runCommentCommand(m *Model, raw string) tea.Cmd {
	m.addComment(raw)
	return m.view.notice.Start("Comment added", "", noticeDuration)
}

func (m *Model) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.view.command.cmd == CmdFilter && msg.String() == "ctrl+p" {
		cfg, err := m.loadFilterConfig()
		if err != nil {
			return m, m.view.notice.Start("Filter presets error", "warn", noticeDuration)
		}
		presets := make([]dialogs.FilterPreset, 0, len(cfg.Presets))
		for _, preset := range cfg.Presets {
			presets = append(presets, dialogs.FilterPreset{
				Pattern:     preset.Pattern,
				Description: preset.Description,
			})
		}
		m.activeDialog = dialogs.NewFilterPaletteDialog(
			presets,
			cfg.History,
			m.terminalWidth,
			m.terminalHeight,
			m.styles.RowSelectedFG,
			m.styles.RowSelectedBG,
		)
		m.activeDialog.Show()
		return m, nil
	}

	// universal cancel
	if msg.Type == tea.KeyEsc {
		cmd := m.exitCommand(true)
		return m, cmd
	}

	// constrained command: mark
	if m.view.command.cmd == CmdMark {
		return m.handleMarkCommandKey(msg) // your tightened function
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

	// editing
	switch msg.Type {
	case tea.KeyBackspace:
		if len(m.view.command.buf) > 0 {
			m.view.command.buf = m.view.command.buf[:len(m.view.command.buf)-1]
		}
		return m, nil
	}

	// append printable rune
	if len(msg.Runes) == 1 {
		m.view.command.buf += string(msg.Runes[0])
	}
	return m, nil
}
