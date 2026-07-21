package dialogs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

type helpLine struct {
	category bool
	text     string
}

type Help struct {
	visible     bool
	items       []CommandItem
	lines       []helpLine
	scroll      int
	width       int
	visibleRows int
}

func NewHelpDialog(items []CommandItem, terminalWidth, terminalHeight int) *Help {
	width := 84
	if available := terminalWidth - 4; terminalWidth > 0 && available < width {
		width = available
	}
	if width < 32 {
		width = 32
	}
	visibleRows := terminalHeight - 7
	if visibleRows < 6 {
		visibleRows = 6
	}
	if visibleRows > 24 {
		visibleRows = 24
	}

	d := &Help{
		visible:     true,
		items:       append([]CommandItem(nil), items...),
		width:       width,
		visibleRows: visibleRows,
	}
	d.lines = buildHelpLines(d.items, max(12, width-4))
	return d
}

func (d Help) Init() tea.Cmd { return nil }

func (d *Help) Update(msg tea.Msg) (Dialog, Action, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter", "esc":
			d.visible = false
			return d, Action{Kind: ActionClose}, nil
		case "up", "k":
			d.move(-1)
		case "down", "j":
			d.move(1)
		case "pgup", "ctrl+u":
			d.move(-d.visibleRows)
		case "pgdown", "ctrl+d":
			d.move(d.visibleRows)
		case "home", "g":
			d.scroll = 0
		case "end", "G":
			d.scroll = d.maxScroll()
		}
	}
	return d, Action{Kind: ActionNone}, nil
}

func (d Help) View() string {
	if !d.visible {
		return ""
	}

	innerWidth := max(12, d.width-4)
	end := min(len(d.lines), d.scroll+d.visibleRows)
	visible := make([]string, 0, d.visibleRows)
	for _, line := range d.lines[d.scroll:end] {
		if line.category {
			visible = append(visible, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(line.text))
		} else {
			visible = append(visible, line.text)
		}
	}
	for len(visible) < d.visibleRows {
		visible = append(visible, "")
	}

	position := "No commands"
	if len(d.lines) > 0 {
		position = fmt.Sprintf("Lines %d-%d of %d", d.scroll+1, end, len(d.lines))
	}
	content := []string{dialogSectionLabel("Command reference")}
	content = append(content, visible...)
	content = append(content,
		"",
		dialogStatusLine("", position+"  |  j/k or up/down: scroll  PgUp/PgDn: page"),
		renderDialogActionRowWithKeys(innerWidth, "Esc", "Close", true, "", ""),
	)

	return renderDialogPanel(
		"Help",
		dialogTopRightState(fmt.Sprintf("%d commands", len(d.items))),
		d.width,
		content,
	)
}

func buildHelpLines(items []CommandItem, width int) []helpLine {
	lines := make([]helpLine, 0, len(items)+12)
	currentCategory := ""
	shortcutWidth := min(20, max(10, width/4))
	for _, item := range items {
		if item.Category != currentCategory {
			if len(lines) > 0 {
				lines = append(lines, helpLine{text: ""})
			}
			currentCategory = item.Category
			lines = append(lines, helpLine{category: true, text: strings.ToUpper(currentCategory)})
		}
		shortcut := xansi.Truncate(item.Shortcut, shortcutWidth, "")
		titleWidth := max(8, width-shortcutWidth-2)
		title := xansi.Truncate(item.Title, titleWidth, "")
		lines = append(lines, helpLine{text: padCommandCell(shortcut, shortcutWidth) + "  " + title})
	}
	return lines
}

func (d *Help) move(delta int) {
	d.scroll += delta
	if d.scroll < 0 {
		d.scroll = 0
	}
	if d.scroll > d.maxScroll() {
		d.scroll = d.maxScroll()
	}
}

func (d Help) maxScroll() int {
	maxScroll := len(d.lines) - d.visibleRows
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func (d *Help) Show()          { d.visible = true }
func (d *Help) Hide()          { d.visible = false }
func (d *Help) Focus() tea.Cmd { return nil }
func (d *Help) Blur()          {}
func (d Help) IsVisible() bool { return d.visible }
