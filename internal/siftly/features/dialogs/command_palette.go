package dialogs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

type CommandPalette struct {
	visible bool
	input   textinput.Model

	items    []CommandItem
	filtered []CommandItem
	cursor   int
	scroll   int

	width       int
	visibleRows int
	selectedFG  lipgloss.Color
	selectedBG  lipgloss.Color
	tokens      ui.DesignTokens
}

func NewCommandPalette(items []CommandItem, terminalWidth, terminalHeight int, selectedFG, selectedBG lipgloss.Color, tokenOptions ...ui.DesignTokens) *CommandPalette {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Search commands, categories or shortcuts"
	input.CharLimit = 256
	tokens := dialogDesignTokens(tokenOptions)
	styleDialogTextInput(&input, tokens)

	d := &CommandPalette{
		visible:    true,
		input:      input,
		items:      append([]CommandItem(nil), items...),
		selectedFG: selectedFG,
		selectedBG: selectedBG,
		tokens:     tokens,
	}
	d.rebuildFiltered()
	d.Resize(terminalWidth, terminalHeight)
	return d
}

func (d CommandPalette) Init() tea.Cmd { return d.input.Focus() }

func (d *CommandPalette) Update(msg tea.Msg) (Dialog, Action, tea.Cmd) {
	if !d.visible {
		return d, Action{Kind: ActionNone}, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			d.visible = false
			return d, Action{Kind: ActionCommandCancel}, nil
		case "enter":
			item, ok := d.selectedItem()
			if !ok || !item.Enabled {
				return d, Action{Kind: ActionNone}, nil
			}
			d.visible = false
			return d, Action{Kind: ActionCommandRun, CommandID: item.ID}, nil
		case "up", "ctrl+p":
			d.move(-1)
			return d, Action{Kind: ActionNone}, nil
		case "down", "ctrl+n":
			d.move(1)
			return d, Action{Kind: ActionNone}, nil
		case "pgup":
			d.move(-d.visibleRows)
			return d, Action{Kind: ActionNone}, nil
		case "pgdown":
			d.move(d.visibleRows)
			return d, Action{Kind: ActionNone}, nil
		}
	}

	previous := d.input.Value()
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	if d.input.Value() != previous {
		d.rebuildFiltered()
	}
	return d, Action{Kind: ActionNone}, cmd
}

func (d CommandPalette) View() string {
	if !d.visible {
		return ""
	}

	innerWidth := max(12, d.width-4)
	selected, hasSelection := d.selectedItem()
	statusKind := "success"
	status := fmt.Sprintf("%d matching commands", len(d.filtered))
	primaryEnabled := hasSelection && selected.Enabled
	if len(d.filtered) == 0 {
		statusKind = "error"
		status = "No matching commands"
	} else if hasSelection && !selected.Enabled {
		statusKind = "warn"
		status = selected.DisabledReason
		if strings.TrimSpace(status) == "" {
			status = "Command is unavailable"
		}
	} else if hasSelection && strings.TrimSpace(selected.Description) != "" {
		status = selected.Description
	}

	content := []string{
		dialogSectionLabel("Search", d.tokens),
		d.input.View(),
		"",
		dialogStatusLine(statusKind, status, d.tokens),
		renderDialogActionRowWithKeys(innerWidth, "Enter", "Run", primaryEnabled, "Esc", "Close", d.tokens),
		"",
		dialogSectionLabel("Commands", d.tokens),
	}
	content = append(content, d.renderRows(innerWidth)...)

	return renderDialogPanel(
		"Command Palette",
		dialogTopRightState(fmt.Sprintf("%d commands", len(d.items)), d.tokens),
		d.width,
		content,
		d.tokens,
	)
}

func (d *CommandPalette) rebuildFiltered() {
	query := strings.ToLower(strings.TrimSpace(d.input.Value()))
	terms := strings.Fields(query)
	d.filtered = d.filtered[:0]
	type scoredCommand struct {
		item  CommandItem
		score int
	}
	scored := make([]scoredCommand, 0, len(d.items))
	for _, item := range d.items {
		score, matches := commandMatchScore(item, query, terms)
		if matches {
			scored = append(scored, scoredCommand{item: item, score: score})
		}
	}
	if query != "" {
		sort.SliceStable(scored, func(i, j int) bool { return scored[i].score > scored[j].score })
	}
	for _, match := range scored {
		d.filtered = append(d.filtered, match.item)
	}
	d.cursor = 0
	d.scroll = 0
	for i, item := range d.filtered {
		if item.Enabled {
			d.cursor = i
			break
		}
	}
}

func commandMatchScore(item CommandItem, query string, terms []string) (int, bool) {
	title := strings.ToLower(item.Title)
	category := strings.ToLower(item.Category)
	shortcut := strings.ToLower(item.Shortcut)
	description := strings.ToLower(item.Description)
	keywords := strings.ToLower(item.Keywords)
	haystack := strings.Join([]string{title, category, shortcut, description, keywords}, " ")
	for _, term := range terms {
		if !strings.Contains(haystack, term) {
			return 0, false
		}
	}
	if query == "" {
		return 0, true
	}

	score := 0
	if shortcut == query {
		score += 300
	}
	if title == query {
		score += 250
	} else if strings.Contains(title, query) {
		score += 120
	}
	if strings.Contains(category, query) {
		score += 80
	}
	if strings.Contains(keywords, query) {
		score += 40
	}
	if strings.Contains(description, query) {
		score += 20
	}
	for _, term := range terms {
		if strings.Contains(title, term) {
			score += 15
		}
		if strings.Contains(category, term) {
			score += 10
		}
		if strings.Contains(keywords, term) {
			score += 5
		}
	}
	return score, true
}

func (d *CommandPalette) move(delta int) {
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

func (d *CommandPalette) ensureCursorVisible() {
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

func (d CommandPalette) selectedItem() (CommandItem, bool) {
	if d.cursor < 0 || d.cursor >= len(d.filtered) {
		return CommandItem{}, false
	}
	return d.filtered[d.cursor], true
}

func (d CommandPalette) renderRows(width int) []string {
	rowCount := min(len(d.filtered), d.visibleRows)
	if rowCount <= 0 {
		return nil
	}
	rows := make([]string, 0, rowCount)
	end := min(len(d.filtered), d.scroll+rowCount)
	for i := d.scroll; i < end; i++ {
		rows = append(rows, d.renderRow(d.filtered[i], width, i == d.cursor))
	}
	return rows
}

func (d *CommandPalette) Resize(terminalWidth, terminalHeight int) {
	d.width = responsiveDialogWidth(terminalWidth, d.preferredWidth(), 64)
	d.input.Width = max(1, d.width-8)
	heightLimit := responsiveDialogHeight(terminalHeight, 25)
	d.visibleRows = boundedListRows(heightLimit, 9, -1, 16)
	d.ensureCursorVisible()
}

func (d CommandPalette) preferredWidth() int {
	innerWidth := lipgloss.Width(d.input.Placeholder)
	for _, item := range d.items {
		rowWidth := lipgloss.Width(item.Category) + lipgloss.Width(item.Title) + lipgloss.Width(item.Shortcut) + 6
		if rowWidth > innerWidth {
			innerWidth = rowWidth
		}
	}
	return clampDialogWidth(innerWidth+4, 64, 88)
}

func (d CommandPalette) renderRow(item CommandItem, width int, selected bool) string {
	categoryWidth := 0
	if width >= 58 {
		categoryWidth = 22
	}
	shortcutWidth := min(18, max(8, width/5))
	separators := 2
	if categoryWidth > 0 {
		separators = 4
	}
	titleWidth := width - categoryWidth - shortcutWidth - separators
	if titleWidth < 8 {
		titleWidth = 8
	}

	title := xansi.Truncate(item.Title, titleWidth, "")
	shortcut := xansi.Truncate(item.Shortcut, shortcutWidth, "")
	line := padCommandCell(title, titleWidth) + "  " + shortcut
	if categoryWidth > 0 {
		category := xansi.Truncate(strings.ToUpper(item.Category), categoryWidth, "")
		line = padCommandCell(category, categoryWidth) + "  " + line
	}
	line = xansi.Truncate(line, width, "")
	line = padCommandCell(line, width)

	tokens := d.tokens
	style := tokens.Emphasis.Normal
	if !item.Enabled {
		style = tokens.Emphasis.Subtle
	}
	if selected {
		style = tokens.States.Selected
		if d.selectedFG != "" {
			style = style.Foreground(d.selectedFG)
		}
		if d.selectedBG != "" {
			style = style.Background(d.selectedBG)
		}
	}
	return style.Render(line)
}

func padCommandCell(value string, width int) string {
	missing := width - xansi.StringWidth(value)
	if missing <= 0 {
		return value
	}
	return value + strings.Repeat(" ", missing)
}

func (d *CommandPalette) Show() {
	d.visible = true
	d.input.Focus()
}

func (d *CommandPalette) Hide() {
	d.visible = false
	d.input.Blur()
}

func (d *CommandPalette) Focus() tea.Cmd { return d.input.Focus() }
func (d *CommandPalette) Blur()          { d.input.Blur() }
func (d CommandPalette) IsVisible() bool { return d.visible }
