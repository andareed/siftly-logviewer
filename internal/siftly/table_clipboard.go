package siftly

import (
	"fmt"

	"github.com/andareed/siftly-hostlog/clipboard"
	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) copyRowToClipboard() tea.Cmd {
	if m.cursor >= 0 && m.cursor < len(m.table.filteredIndices) {
		row := m.table.rows[m.table.filteredIndices[m.cursor]]
		text := row.Join("\t") // Tab delimited string
		if err := clipboard.Copy(text); err != nil {
			logging.Errorf("Clipboard copy failed: %v", err)
			return m.view.notice.Start(fmt.Sprintf("Clipboard error: %v", err), "warn", noticeDuration)
		}
		return m.view.notice.Start("Copied row to clipboard", "", noticeDuration)
	}
	return nil
}
