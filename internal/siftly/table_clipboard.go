package siftly

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/clipboard"
	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxClipboardRows  = 10_000
	maxClipboardBytes = 1 << 20
)

func (m *Model) copyRowsToClipboard() tea.Cmd {
	text, count, err := m.selectedRowsClipboardText()
	if err != nil {
		return m.view.notice.Start(err.Error(), "warn", noticeDuration)
	}
	if count == 0 {
		return nil
	}
	if err := clipboard.Copy(text); err != nil {
		logging.Errorf("Clipboard copy failed: %v", err)
		return m.view.notice.Start(fmt.Sprintf("Clipboard error: %v", err), "warn", noticeDuration)
	}
	if count == 1 {
		return m.view.notice.Start("Copied row to clipboard", "", noticeDuration)
	}
	return m.view.notice.Start(fmt.Sprintf("Copied %d rows to clipboard", count), "", noticeDuration)
}

func (m *Model) copyInspectorFieldToClipboard() tea.Cmd {
	name, value, ok := m.inspectorFieldClipboardText()
	if !ok {
		return m.view.notice.Start("No field to copy", "warn", noticeDuration)
	}
	if len(value) > maxClipboardBytes {
		return m.view.notice.Start(
			fmt.Sprintf("Field exceeds the %d KiB clipboard limit; export instead", maxClipboardBytes/1024),
			"warn",
			noticeDuration,
		)
	}
	if err := clipboard.Copy(value); err != nil {
		logging.Errorf("Clipboard field copy failed: %v", err)
		return m.view.notice.Start(fmt.Sprintf("Clipboard error: %v", err), "warn", noticeDuration)
	}
	return m.view.notice.Start(fmt.Sprintf("Copied %s field", name), "", noticeDuration)
}

func (m *Model) inspectorFieldClipboardText() (string, string, bool) {
	row, _, ok := m.currentInspectorRow()
	if !ok || len(m.table.header) == 0 {
		return "", "", false
	}
	field := m.inspectorFieldIndex()
	column := m.table.header[field]
	return inspectorColumnLabel(column, field), inspectorColumnValue(row, column, field), true
}

func (m *Model) selectedRowsClipboardText() (string, int, error) {
	start, end, ok := m.actionDisplayRange()
	if !ok {
		return "", 0, nil
	}
	count := end - start + 1
	if count > maxClipboardRows {
		return "", 0, fmt.Errorf("Selection too large for clipboard (%d rows); export instead", count)
	}

	var b strings.Builder
	for displayIndex := start; displayIndex <= end; displayIndex++ {
		rowIndex := m.table.filteredIndices[displayIndex]
		line := m.table.rows[rowIndex].Join("\t")
		separatorBytes := 0
		if displayIndex > start {
			separatorBytes = 1
		}
		if b.Len()+separatorBytes+len(line) > maxClipboardBytes {
			return "", 0, fmt.Errorf("Selection exceeds the %d KiB clipboard limit; export instead", maxClipboardBytes/1024)
		}
		if separatorBytes > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}
	return b.String(), count, nil
}
