package siftly

import (
	"fmt"
)

func (m *Model) commandPreviewSuffix() string {
	switch m.view.command.cmd {
	case CmdSort:
		if _, err := parseSortSpec(m.view.command.buf, m.table.header); err != nil {
			return " (invalid sort)"
		}
		return fmt.Sprintf(" (%d matches)", len(m.table.filteredIndices))
	default:
		return ""
	}
}
