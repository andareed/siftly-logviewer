package siftly

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

type sortSpec struct {
	Enabled     bool
	ColumnIndex int
	Desc        bool
}

const sortColumnOriginalIndex = -2

func (m *Model) setSortSpec(input string) error {
	spec, err := parseSortSpec(input, m.table.header)
	if err != nil {
		return err
	}

	m.table.sortEnabled = spec.Enabled
	if spec.Enabled {
		m.table.sortColumn = spec.ColumnIndex
		m.table.sortDesc = spec.Desc
	} else {
		m.table.sortColumn = -1
		m.table.sortDesc = false
	}

	m.rebuildRowOrder()
	m.applyFilter()
	return nil
}

func (m *Model) currentSortSeed() string {
	if !m.table.sortEnabled {
		return ""
	}
	colName, ok := m.sortColumnName(m.table.sortColumn)
	if !ok {
		return ""
	}
	if strings.ContainsAny(colName, " \t") {
		colName = `"` + strings.ReplaceAll(colName, `"`, `\"`) + `"`
	}
	direction := "asc"
	if m.table.sortDesc {
		direction = "desc"
	}
	return fmt.Sprintf("%s %s", colName, direction)
}

func (m *Model) sortedHeaderName(col ui.ColumnMeta) string {
	name := col.Name
	if !m.table.sortEnabled || col.Index != m.table.sortColumn {
		return name
	}
	if m.table.sortDesc {
		return name + " ↓"
	}
	return name + " ↑"
}

func (m *Model) rebuildRowOrder() {
	defer logging.TimeOperation("rebuild row order")()

	m.table.rowOrder = makeIndexRange(len(m.table.rows))
	if !m.table.sortEnabled || len(m.table.rowOrder) < 2 {
		return
	}

	column := m.table.sortColumn
	desc := m.table.sortDesc
	sort.SliceStable(m.table.rowOrder, func(i, j int) bool {
		leftRow := m.table.rowOrder[i]
		rightRow := m.table.rowOrder[j]
		cmp := m.compareRowsByColumn(leftRow, rightRow, column)
		if cmp == 0 {
			return false
		}
		if desc {
			return cmp > 0
		}
		return cmp < 0
	})
}

func (m *Model) compareRowsByColumn(leftRow, rightRow, column int) int {
	if column == sortColumnOriginalIndex {
		left := m.table.rows[leftRow].OriginalIndex
		right := m.table.rows[rightRow].OriginalIndex
		switch {
		case left < right:
			return -1
		case left > right:
			return 1
		default:
			return 0
		}
	}

	left := columnValue(m.table.rows[leftRow].Cols, column)
	right := columnValue(m.table.rows[rightRow].Cols, column)
	return compareColumnValues(left, right)
}

func columnValue(cols []string, column int) string {
	if column < 0 || column >= len(cols) {
		return ""
	}
	return strings.TrimSpace(cols[column])
}

func compareColumnValues(left, right string) int {
	if lNum, ok := parseSortNumber(left); ok {
		if rNum, ok := parseSortNumber(right); ok {
			switch {
			case lNum < rNum:
				return -1
			case lNum > rNum:
				return 1
			default:
				return 0
			}
		}
	}

	if lTime, ok := featuretimewindow.ParseLogTimestamp(left); ok {
		if rTime, ok := featuretimewindow.ParseLogTimestamp(right); ok {
			switch {
			case lTime.Before(rTime):
				return -1
			case lTime.After(rTime):
				return 1
			default:
				return 0
			}
		}
	}

	lFold := strings.ToLower(strings.TrimSpace(left))
	rFold := strings.ToLower(strings.TrimSpace(right))
	switch {
	case lFold < rFold:
		return -1
	case lFold > rFold:
		return 1
	}

	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func parseSortNumber(raw string) (float64, bool) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, ",", ""))
	if raw == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func parseSortSpec(raw string, header []ui.ColumnMeta) (sortSpec, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return sortSpec{Enabled: false}, nil
	}
	if strings.EqualFold(raw, "off") || strings.EqualFold(raw, "none") || strings.EqualFold(raw, "clear") {
		return sortSpec{Enabled: false}, nil
	}

	tokens, err := tokenizeSortInput(raw)
	if err != nil {
		return sortSpec{}, err
	}
	if len(tokens) == 0 {
		return sortSpec{}, fmt.Errorf("empty sort expression")
	}

	desc := false
	last := strings.ToLower(tokens[len(tokens)-1])
	if last == "asc" || last == "desc" {
		desc = last == "desc"
		tokens = tokens[:len(tokens)-1]
	}
	if len(tokens) == 0 {
		return sortSpec{}, fmt.Errorf("missing column name")
	}

	columnExpr := strings.TrimSpace(strings.Join(tokens, " "))
	columnIdx, err := resolveSortColumn(columnExpr, header)
	if err != nil {
		return sortSpec{}, err
	}

	return sortSpec{
		Enabled:     true,
		ColumnIndex: columnIdx,
		Desc:        desc,
	}, nil
}

func tokenizeSortInput(raw string) ([]string, error) {
	tokens := make([]string, 0, 4)
	n := len(raw)
	for i := 0; i < n; {
		for i < n && (raw[i] == ' ' || raw[i] == '\t' || raw[i] == '\n' || raw[i] == '\r') {
			i++
		}
		if i >= n {
			break
		}

		if raw[i] == '"' || raw[i] == '\'' {
			quote := raw[i]
			i++
			var b strings.Builder
			closed := false
			for i < n {
				c := raw[i]
				if c == '\\' && i+1 < n && raw[i+1] == quote {
					b.WriteByte(raw[i+1])
					i += 2
					continue
				}
				if c == quote {
					i++
					closed = true
					break
				}
				b.WriteByte(c)
				i++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated quote")
			}
			tokens = append(tokens, b.String())
			continue
		}

		start := i
		for i < n && raw[i] != ' ' && raw[i] != '\t' && raw[i] != '\n' && raw[i] != '\r' {
			i++
		}
		tokens = append(tokens, raw[start:i])
	}

	return tokens, nil
}

func resolveSortColumn(columnExpr string, header []ui.ColumnMeta) (int, error) {
	columnExpr = strings.TrimSpace(columnExpr)
	if columnExpr == "" {
		return -1, fmt.Errorf("missing column name")
	}

	if isOriginalIndexSortAlias(columnExpr) {
		return sortColumnOriginalIndex, nil
	}

	if n, err := strconv.Atoi(columnExpr); err == nil {
		return resolveVisibleSortColumnByNumber(n, header)
	}

	needle := strings.TrimSpace(strings.TrimPrefix(columnExpr, "\ufeff"))
	for _, col := range header {
		name := strings.TrimSpace(strings.TrimPrefix(col.Name, "\ufeff"))
		if strings.EqualFold(name, needle) {
			return col.Index, nil
		}
	}

	return -1, fmt.Errorf("unknown column %q", columnExpr)
}

func resolveVisibleSortColumnByNumber(n int, header []ui.ColumnMeta) (int, error) {
	if n == 0 {
		return sortColumnOriginalIndex, nil
	}
	if n < 0 {
		return -1, fmt.Errorf("column number must be >= 0")
	}

	visibleNum := 0
	for _, col := range header {
		if !col.Visible {
			continue
		}
		visibleNum++
		if visibleNum == n {
			return col.Index, nil
		}
	}

	return -1, fmt.Errorf("column number %d out of range", n)
}

func (m *Model) sortColumnName(column int) (string, bool) {
	if column == sortColumnOriginalIndex {
		return "row", true
	}
	for _, col := range m.table.header {
		if col.Index == column {
			return col.Name, true
		}
	}
	return "", false
}

func isOriginalIndexSortAlias(expr string) bool {
	expr = strings.TrimSpace(strings.TrimPrefix(expr, "\ufeff"))
	switch strings.ToLower(expr) {
	case "#", "row", "rownum", "rownumber", "line", "linenumber", "index":
		return true
	default:
		return false
	}
}
