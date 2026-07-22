package dialogs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

type ColumnManager struct {
	visible bool
	input   textinput.Model

	columns  []ColumnManagerItem
	defaults []ColumnManagerItem
	filtered []int
	cursor   int
	scroll   int

	sortEnabled bool
	sortColumn  int
	sortDesc    bool
	searchFocus bool
	statusKind  string
	status      string

	width       int
	visibleRows int
	selectedFG  lipgloss.Color
	selectedBG  lipgloss.Color
}

func NewColumnManager(
	columns []ColumnManagerItem,
	defaults []ColumnManagerItem,
	sortEnabled bool,
	sortColumn int,
	sortDesc bool,
	terminalWidth int,
	terminalHeight int,
	selectedFG lipgloss.Color,
	selectedBG lipgloss.Color,
) *ColumnManager {
	width := 88
	if available := terminalWidth - 4; terminalWidth > 0 && available < width {
		width = available
	}
	if width < 42 {
		width = 42
	}

	visibleRows := terminalHeight - 13
	if visibleRows < 5 {
		visibleRows = 5
	}
	if visibleRows > 16 {
		visibleRows = 16
	}

	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Press / to search columns"
	input.CharLimit = 256
	input.Width = max(12, width-8)
	input.Blur()

	d := &ColumnManager{
		visible:     true,
		input:       input,
		columns:     append([]ColumnManagerItem(nil), columns...),
		defaults:    append([]ColumnManagerItem(nil), defaults...),
		sortEnabled: sortEnabled,
		sortColumn:  sortColumn,
		sortDesc:    sortDesc,
		width:       width,
		visibleRows: visibleRows,
		selectedFG:  selectedFG,
		selectedBG:  selectedBG,
	}
	if len(d.defaults) != len(d.columns) {
		d.defaults = append([]ColumnManagerItem(nil), d.columns...)
	}
	d.rebuildFiltered(-1)
	return d
}

func (d ColumnManager) Init() tea.Cmd { return nil }

func (d *ColumnManager) Update(msg tea.Msg) (Dialog, Action, tea.Cmd) {
	if !d.visible {
		return d, Action{Kind: ActionNone}, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if d.searchFocus {
			return d.updateSearch(keyMsg)
		}
		return d.updateList(keyMsg)
	}
	return d, Action{Kind: ActionNone}, nil
}

func (d *ColumnManager) updateSearch(msg tea.KeyMsg) (Dialog, Action, tea.Cmd) {
	switch msg.String() {
	case "esc":
		d.visible = false
		return d, Action{Kind: ActionColumnManagerCancel}, nil
	case "tab", "enter":
		d.searchFocus = false
		d.input.Blur()
		return d, Action{Kind: ActionNone}, nil
	case "up", "ctrl+p":
		d.move(-1)
		return d, Action{Kind: ActionNone}, nil
	case "down", "ctrl+n":
		d.move(1)
		return d, Action{Kind: ActionNone}, nil
	}

	selected := d.selectedSourceIndex()
	previous := d.input.Value()
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	if d.input.Value() != previous {
		d.rebuildFiltered(selected)
	}
	return d, Action{Kind: ActionNone}, cmd
}

func (d *ColumnManager) updateList(msg tea.KeyMsg) (Dialog, Action, tea.Cmd) {
	switch msg.String() {
	case "esc":
		d.visible = false
		return d, Action{Kind: ActionColumnManagerCancel}, nil
	case "enter":
		d.visible = false
		result := d.result()
		return d, Action{Kind: ActionColumnManagerApply, ColumnManager: &result}, nil
	case "/":
		d.searchFocus = true
		d.input.Placeholder = "Search columns"
		return d, Action{Kind: ActionNone}, d.input.Focus()
	case "up", "k":
		d.move(-1)
	case "down", "j":
		d.move(1)
	case "pgup":
		d.move(-d.visibleRows)
	case "pgdown":
		d.move(d.visibleRows)
	case "ctrl+up":
		d.reorderSelected(-1)
	case "ctrl+down":
		d.reorderSelected(1)
	case "K":
		d.reorderSelected(-1)
	case "J":
		d.reorderSelected(1)
	case " ":
		d.toggleVisibility()
	case "s":
		d.cycleSort()
	case "f":
		d.toggleFrozen()
	case "a":
		d.queueAutoFit(false)
	case "A":
		d.queueAutoFit(true)
	case "r":
		d.resetDraft()
	case "tab":
		d.searchFocus = true
		return d, Action{Kind: ActionNone}, d.input.Focus()
	}
	return d, Action{Kind: ActionNone}, nil
}

func (d ColumnManager) View() string {
	if !d.visible {
		return ""
	}

	innerWidth := max(12, d.width-4)
	statusKind := d.statusKind
	status := d.status
	if len(d.filtered) == 0 {
		statusKind = "warn"
		status = "No columns match the search"
	} else if status == "" {
		visible, frozen := d.columnCounts()
		status = fmt.Sprintf("%d visible, %d frozen", visible, frozen)
		if sort := d.sortStatus(); sort != "" {
			status += ", " + sort
		}
	}

	content := []string{
		dialogSectionLabel("Search"),
		d.input.View(),
		"",
		dialogStatusLine(statusKind, status),
		renderDialogActionRowWithKeys(innerWidth, "Enter", "Apply", len(d.columns) > 0, "Esc", "Cancel"),
		"",
		dialogSectionLabel("Columns"),
	}
	content = append(content, d.renderRows(innerWidth)...)
	content = append(content, "", lipgloss.NewStyle().Faint(true).Render(
		"Space show/hide   s sort   f freeze   a/A auto-fit   J/K move   r reset   / search",
	))

	return renderDialogPanel(
		"Column Manager",
		dialogTopRightState(fmt.Sprintf("%d columns", len(d.columns))),
		d.width,
		content,
	)
}

func (d *ColumnManager) toggleVisibility() {
	index, ok := d.selectedColumnIndex()
	if !ok {
		return
	}
	if d.columns[index].Visible && d.visibleCount() == 1 {
		d.setStatus("warn", "At least one column must remain visible")
		return
	}
	d.columns[index].Visible = !d.columns[index].Visible
	if !d.columns[index].Visible {
		d.columns[index].Frozen = false
		selectedSource := d.columns[index].SourceIndex
		d.normalizeFrozenOrder()
		d.rebuildFiltered(selectedSource)
		index, _ = d.selectedColumnIndex()
	}
	d.setStatus("", fmt.Sprintf("%s %s", d.columns[index].Name, visibilityLabel(d.columns[index].Visible)))
}

func (d *ColumnManager) cycleSort() {
	index, ok := d.selectedColumnIndex()
	if !ok {
		return
	}
	sourceIndex := d.columns[index].SourceIndex
	switch {
	case !d.sortEnabled || d.sortColumn != sourceIndex:
		d.sortEnabled = true
		d.sortColumn = sourceIndex
		d.sortDesc = false
		d.setStatus("", "Sort "+d.columns[index].Name+" ascending")
	case !d.sortDesc:
		d.sortDesc = true
		d.setStatus("", "Sort "+d.columns[index].Name+" descending")
	default:
		d.sortEnabled = false
		d.sortColumn = -1
		d.sortDesc = false
		d.setStatus("", "Sorting cleared")
	}
}

func (d *ColumnManager) toggleFrozen() {
	index, ok := d.selectedColumnIndex()
	if !ok {
		return
	}
	selectedSource := d.columns[index].SourceIndex
	if !d.columns[index].Visible {
		d.columns[index].Visible = true
	}
	d.columns[index].Frozen = !d.columns[index].Frozen
	frozen := d.columns[index].Frozen
	d.normalizeFrozenOrder()
	d.rebuildFiltered(selectedSource)
	state := "unfrozen"
	if frozen {
		state = "frozen"
	}
	d.setStatus("", d.columnName(selectedSource)+" "+state)
}

func (d *ColumnManager) queueAutoFit(all bool) {
	if all {
		count := 0
		for i := range d.columns {
			if d.columns[i].Visible {
				d.columns[i].AutoFit = true
				count++
			}
		}
		d.setStatus("", fmt.Sprintf("Auto-fit queued for %d visible columns", count))
		return
	}
	index, ok := d.selectedColumnIndex()
	if !ok {
		return
	}
	d.columns[index].AutoFit = true
	d.setStatus("", "Auto-fit queued for "+d.columns[index].Name)
}

func (d *ColumnManager) resetDraft() {
	selectedSource := d.selectedSourceIndex()
	d.columns = append(d.columns[:0], d.defaults...)
	for i := range d.columns {
		d.columns[i].AutoFit = false
	}
	d.sortEnabled = false
	d.sortColumn = -1
	d.sortDesc = false
	d.input.SetValue("")
	d.input.Blur()
	d.searchFocus = false
	d.rebuildFiltered(selectedSource)
	d.setStatus("", "Layout reset queued; Enter to apply or Esc to cancel")
}

func (d *ColumnManager) reorderSelected(delta int) {
	index, ok := d.selectedColumnIndex()
	if !ok || delta == 0 {
		return
	}
	target := index + delta
	if target < 0 || target >= len(d.columns) {
		return
	}
	if d.columns[index].Frozen != d.columns[target].Frozen {
		d.setStatus("warn", "Frozen columns stay before scrolling columns")
		return
	}
	selectedSource := d.columns[index].SourceIndex
	d.columns[index], d.columns[target] = d.columns[target], d.columns[index]
	d.rebuildFiltered(selectedSource)
	d.setStatus("", "Moved "+d.columnName(selectedSource))
}

func (d *ColumnManager) normalizeFrozenOrder() {
	frozen := make([]ColumnManagerItem, 0, len(d.columns))
	unfrozen := make([]ColumnManagerItem, 0, len(d.columns))
	for _, column := range d.columns {
		if column.Frozen {
			frozen = append(frozen, column)
		} else {
			unfrozen = append(unfrozen, column)
		}
	}
	d.columns = append(frozen, unfrozen...)
}

func (d *ColumnManager) rebuildFiltered(selectedSource int) {
	query := strings.ToLower(strings.TrimSpace(d.input.Value()))
	d.filtered = d.filtered[:0]
	for i, column := range d.columns {
		if query == "" || strings.Contains(strings.ToLower(column.Name), query) {
			d.filtered = append(d.filtered, i)
		}
	}
	d.cursor = 0
	if selectedSource >= 0 {
		for i, columnIndex := range d.filtered {
			if d.columns[columnIndex].SourceIndex == selectedSource {
				d.cursor = i
				break
			}
		}
	}
	d.scroll = 0
	d.ensureCursorVisible()
}

func (d *ColumnManager) move(delta int) {
	if len(d.filtered) == 0 {
		return
	}
	d.cursor += delta
	if d.cursor < 0 {
		d.cursor = 0
	}
	if d.cursor >= len(d.filtered) {
		d.cursor = len(d.filtered) - 1
	}
	d.ensureCursorVisible()
}

func (d *ColumnManager) ensureCursorVisible() {
	if d.cursor < d.scroll {
		d.scroll = d.cursor
	}
	if d.cursor >= d.scroll+d.visibleRows {
		d.scroll = d.cursor - d.visibleRows + 1
	}
	maxScroll := len(d.filtered) - d.visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if d.scroll > maxScroll {
		d.scroll = maxScroll
	}
}

func (d ColumnManager) renderRows(width int) []string {
	rows := make([]string, 0, d.visibleRows)
	end := min(len(d.filtered), d.scroll+d.visibleRows)
	for i := d.scroll; i < end; i++ {
		columnIndex := d.filtered[i]
		rows = append(rows, d.renderRow(d.columns[columnIndex], columnIndex, width, i == d.cursor))
	}
	for len(rows) < d.visibleRows {
		rows = append(rows, "")
	}
	return rows
}

func (d ColumnManager) renderRow(column ColumnManagerItem, columnIndex int, width int, selected bool) string {
	position := fmt.Sprintf("%2d", columnIndex+1)
	check := "[ ]"
	if column.Visible {
		check = "[x]"
	}

	states := make([]string, 0, 3)
	if column.Frozen {
		states = append(states, "FROZEN")
	}
	if d.sortEnabled && d.sortColumn == column.SourceIndex {
		direction := "ASC"
		if d.sortDesc {
			direction = "DESC"
		}
		states = append(states, "SORT "+direction)
	}
	if column.AutoFit {
		states = append(states, "AUTO")
	} else {
		states = append(states, fmt.Sprintf("%d COLS", column.MinWidth))
	}

	right := strings.Join(states, "  ")
	prefix := position + "  " + check + "  "
	nameWidth := width - xansi.StringWidth(prefix) - xansi.StringWidth(right) - 2
	if nameWidth < 6 {
		nameWidth = 6
	}
	name := xansi.Truncate(column.Name, nameWidth, "")
	line := prefix + padCommandCell(name, nameWidth) + "  " + right
	line = padCommandCell(xansi.Truncate(line, width, ""), width)

	style := lipgloss.NewStyle()
	if !column.Visible {
		style = style.Faint(true)
	}
	if selected {
		style = style.Foreground(d.selectedFG).Background(d.selectedBG)
	}
	return style.Render(line)
}

func (d ColumnManager) result() ColumnManagerResult {
	return ColumnManagerResult{
		Columns:     append([]ColumnManagerItem(nil), d.columns...),
		SortEnabled: d.sortEnabled,
		SortColumn:  d.sortColumn,
		SortDesc:    d.sortDesc,
	}
}

func (d ColumnManager) selectedColumnIndex() (int, bool) {
	if d.cursor < 0 || d.cursor >= len(d.filtered) {
		return -1, false
	}
	index := d.filtered[d.cursor]
	return index, index >= 0 && index < len(d.columns)
}

func (d ColumnManager) selectedSourceIndex() int {
	index, ok := d.selectedColumnIndex()
	if !ok {
		return -1
	}
	return d.columns[index].SourceIndex
}

func (d ColumnManager) visibleCount() int {
	count := 0
	for _, column := range d.columns {
		if column.Visible {
			count++
		}
	}
	return count
}

func (d ColumnManager) columnCounts() (visible int, frozen int) {
	for _, column := range d.columns {
		if column.Visible {
			visible++
		}
		if column.Visible && column.Frozen {
			frozen++
		}
	}
	return visible, frozen
}

func (d ColumnManager) columnName(sourceIndex int) string {
	for _, column := range d.columns {
		if column.SourceIndex == sourceIndex {
			return column.Name
		}
	}
	return "Column"
}

func (d ColumnManager) sortStatus() string {
	if !d.sortEnabled {
		return ""
	}
	name := "source row"
	for _, column := range d.columns {
		if column.SourceIndex == d.sortColumn {
			name = column.Name
			break
		}
	}
	direction := "ascending"
	if d.sortDesc {
		direction = "descending"
	}
	return "sort " + name + " " + direction
}

func (d *ColumnManager) setStatus(kind, status string) {
	d.statusKind = kind
	d.status = status
}

func visibilityLabel(visible bool) string {
	if visible {
		return "shown"
	}
	return "hidden"
}

func (d *ColumnManager) Show() {
	d.visible = true
}

func (d *ColumnManager) Hide() {
	d.visible = false
	d.input.Blur()
}

func (d *ColumnManager) Focus() tea.Cmd {
	if d.searchFocus {
		return d.input.Focus()
	}
	return nil
}

func (d *ColumnManager) Blur()          { d.input.Blur() }
func (d ColumnManager) IsVisible() bool { return d.visible }
