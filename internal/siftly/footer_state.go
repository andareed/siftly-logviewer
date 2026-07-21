package siftly

import (
	"fmt"
	"strings"
)

func (m *Model) footerActiveStates() []string {
	states := make([]string, 0, 6)
	if m.table.filterEnabled && m.table.filterRegex != nil {
		states = append(states, "FILTER")
	}
	if m.table.sortEnabled {
		name, ok := m.sortColumnName(m.table.sortColumn)
		if !ok {
			name = "ROWS"
		}
		direction := "↑"
		if m.table.sortDesc {
			direction = "↓"
		}
		states = append(states, "SORT "+name+" "+direction)
	}
	if m.table.timeWindow.Enabled {
		states = append(states, "TIME WINDOW")
	}
	if m.table.showOnlyMarked {
		states = append(states, "MARKS ONLY")
	}
	if count := m.selectedRowCount(); count > 0 {
		states = append(states, fmt.Sprintf("%d SELECTED", count))
	}
	if m.dirty {
		states = append(states, "UNSAVED")
	}
	return states
}

func (m *Model) footerHints(inputMode bool, command Command) string {
	if inputMode {
		if m.view.mode == modeTimeWindow {
			return "tab: next   enter: apply   r: reset   esc: cancel   ←/→: adjust"
		}
		return m.commandHintsLine(command)
	}

	if prefix := strings.TrimSpace(m.view.pendingViewPrefix); prefix != "" {
		switch prefix {
		case "v":
			return "c: columns   s: sort   o: order   r: reset   esc: cancel"
		case "c":
			return "e: edit comment   v: toggle drawer   esc: cancel"
		case "t":
			return "w: window   b: set start   e: set end   r: reset   esc: cancel"
		}
	}
	if m.view.inspector.open {
		return "j/k: rows   tab/shift+tab: fields   J/K: scroll   y: copy field   ctrl+c: copy row   enter: close"
	}
	if m.selectedRowCount() > 0 {
		return "j/k: extend   ctrl+c: copy   m: mark   space/esc: clear   ?: help"
	}

	hints := "j/k: navigate   f: filter   /: search   space: select   v/c/t: actions   ?: help"
	if m.graphConfig.Enabled {
		hints += "   w: graph"
	}
	return hints
}

func (m *Model) markDirty() {
	m.dirty = true
}
