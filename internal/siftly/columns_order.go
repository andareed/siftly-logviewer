package siftly

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// reorderColumnsBySpec reorders table columns according to spec.
// Spec accepts comma/space separated column names or 1-based positions.
// Unspecified columns are appended in their existing order.
func (m *Model) reorderColumnsBySpec(spec string) (ordered []string, missing []string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil, fmt.Errorf("no columns specified")
	}

	tokens, err := parseColumnTokens(spec)
	if err != nil {
		return nil, nil, err
	}
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("no columns specified")
	}

	selected := make([]int, 0, len(tokens))
	seen := make(map[int]struct{}, len(tokens))
	for _, token := range tokens {
		idx, ok := m.resolveColumnOrderIndex(token)
		if !ok {
			missing = append(missing, token)
			continue
		}
		if _, dup := seen[idx]; dup {
			continue
		}
		seen[idx] = struct{}{}
		selected = append(selected, idx)
		ordered = append(ordered, m.table.header[idx].Name)
	}

	if len(selected) == 0 {
		return nil, missing, fmt.Errorf("no valid columns in order spec")
	}

	newHeader := make([]ui.ColumnMeta, 0, len(m.table.header))
	for _, idx := range selected {
		newHeader = append(newHeader, m.table.header[idx])
	}
	for i := range m.table.header {
		if _, ok := seen[i]; ok {
			continue
		}
		newHeader = append(newHeader, m.table.header[i])
	}

	m.table.header = newHeader
	m.refreshView("reorder-columns", true)
	return ordered, missing, nil
}

func (m *Model) resolveColumnOrderIndex(token string) (int, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return -1, false
	}

	if n, err := strconv.Atoi(token); err == nil {
		n--
		if n >= 0 && n < len(m.table.header) {
			return n, true
		}
	}

	needle := strings.ToLower(token)
	for i, col := range m.table.header {
		if strings.ToLower(col.Name) == needle {
			return i, true
		}
	}
	return -1, false
}

func (m *Model) currentColumnOrderSeed() string {
	parts := make([]string, 0, len(m.table.header))
	for _, col := range m.table.header {
		name := col.Name
		if strings.ContainsAny(name, " \t,") {
			name = `"` + strings.ReplaceAll(name, `"`, `\\"`) + `"`
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ", ")
}

func (m *Model) resetViewLayoutState() tea.Cmd {
	sort.SliceStable(m.table.header, func(i, j int) bool {
		return m.table.header[i].Index < m.table.header[j].Index
	})
	for i := range m.table.header {
		m.table.header[i].Visible = true
	}
	m.table.sortEnabled = false
	m.table.sortColumn = -1
	m.table.sortDesc = false
	m.viewport.ScrollLeft(1 << 20)
	m.refreshView("reset-view-layout", true)
	return m.view.notice.Start("Reset view: visibility, sort, order", "", noticeDuration)
}
