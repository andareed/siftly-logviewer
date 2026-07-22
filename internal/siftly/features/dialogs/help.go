package dialogs

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

type helpLine struct {
	category      bool
	text          string
	command       CommandItem
	commandNumber int
}

type helpColumnWidths struct {
	shortcut   int
	title      int
	detail     int
	showDetail bool
}

type Help struct {
	visible     bool
	items       []CommandItem
	lines       []helpLine
	scroll      int
	width       int
	visibleRows int
	tokens      ui.DesignTokens
}

func NewHelpDialog(items []CommandItem, terminalWidth, terminalHeight int, tokenOptions ...ui.DesignTokens) *Help {
	d := &Help{
		visible: true,
		items:   append([]CommandItem(nil), items...),
		tokens:  dialogDesignTokens(tokenOptions),
	}
	d.Resize(terminalWidth, terminalHeight)
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
	tokens := d.tokens
	for _, line := range d.lines[d.scroll:end] {
		if line.category {
			visible = append(visible, tokens.States.Accent.Render(strings.ToUpper(line.text)))
		} else if line.commandNumber > 0 {
			visible = append(visible, renderHelpCommand(line.command, innerWidth, tokens))
		} else {
			visible = append(visible, line.text)
		}
	}
	position := "No commands"
	if len(d.lines) > 0 {
		first, last := visibleHelpCommandRange(d.lines[d.scroll:end])
		if first > 0 {
			position = fmt.Sprintf("Commands %d-%d of %d", first, last, len(d.items))
		} else {
			position = fmt.Sprintf("%d commands", len(d.items))
		}
	}
	content := []string{renderHelpColumnHeader(innerWidth, d.tokens)}
	content = append(content, visible...)
	content = append(content,
		"",
		dialogStatusLine("", position+"   j/k: move   PgUp/PgDn: page   g/G: first/last", d.tokens),
		renderDialogActionRowWithKeys(innerWidth, "Esc", "Close", true, "", "", d.tokens),
	)

	return renderDialogPanel(
		"Keyboard Reference",
		dialogTopRightState(fmt.Sprintf("%d commands", len(d.items)), d.tokens),
		d.width,
		content,
		d.tokens,
	)
}

func buildHelpLines(items []CommandItem) []helpLine {
	lines := make([]helpLine, 0, len(items)+12)
	currentCategory := ""
	for i, item := range items {
		if item.Category != currentCategory {
			if len(lines) > 0 {
				lines = append(lines, helpLine{text: ""})
			}
			currentCategory = item.Category
			lines = append(lines, helpLine{category: true, text: currentCategory})
		}
		lines = append(lines, helpLine{command: item, commandNumber: i + 1})
	}
	return lines
}

func renderHelpColumnHeader(width int, tokens ui.DesignTokens) string {
	columns := calculateHelpColumnWidths(width)
	style := tokens.Emphasis.Muted.Copy().Bold(true)
	line := style.Render(padCommandCell("KEY", columns.shortcut)) + "  " +
		style.Render(padCommandCell("ACTION", columns.title))
	if columns.showDetail {
		line += "  " + style.Render(padCommandCell("DETAIL", columns.detail))
	}
	return xansi.Truncate(line, width, "")
}

func renderHelpCommand(item CommandItem, width int, tokens ui.DesignTokens) string {
	columns := calculateHelpColumnWidths(width)
	shortcut := padCommandCell(xansi.Truncate(item.Shortcut, columns.shortcut, ""), columns.shortcut)
	title := padCommandCell(xansi.Truncate(item.Title, columns.title, "…"), columns.title)

	shortcutStyle := tokens.States.Accent
	titleStyle := tokens.Emphasis.Normal
	detailStyle := tokens.Emphasis.Muted
	if !item.Enabled {
		shortcutStyle = tokens.Emphasis.Subtle
		titleStyle = tokens.Emphasis.Subtle
		detailStyle = tokens.Emphasis.Subtle
	}

	line := shortcutStyle.Render(shortcut) + "  " + titleStyle.Render(title)
	if columns.showDetail {
		detail := padCommandCell(xansi.Truncate(item.Description, columns.detail, "…"), columns.detail)
		line += "  " + detailStyle.Render(detail)
	}
	return xansi.Truncate(line, width, "")
}

func calculateHelpColumnWidths(width int) helpColumnWidths {
	if width >= 68 {
		shortcut := 18
		detail := max(20, width*2/5)
		title := width - shortcut - detail - 4
		return helpColumnWidths{shortcut: shortcut, title: max(8, title), detail: detail, showDetail: true}
	}
	shortcut := min(18, max(6, width/4))
	title := max(1, width-shortcut-2)
	return helpColumnWidths{shortcut: shortcut, title: title}
}

func visibleHelpCommandRange(lines []helpLine) (first, last int) {
	for _, line := range lines {
		if line.commandNumber == 0 {
			continue
		}
		if first == 0 {
			first = line.commandNumber
		}
		last = line.commandNumber
	}
	return first, last
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

func (d *Help) Resize(terminalWidth, terminalHeight int) {
	d.width = responsiveDialogWidth(terminalWidth, d.preferredWidth(), 40)
	d.lines = buildHelpLines(d.items)
	heightLimit := responsiveDialogHeight(terminalHeight, 30)
	d.visibleRows = boundedListRows(heightLimit, 6, len(d.lines), 24)
	if d.scroll > d.maxScroll() {
		d.scroll = d.maxScroll()
	}
}

func (d Help) preferredWidth() int {
	innerWidth := lipgloss.Width("Keyboard Reference")
	for _, item := range d.items {
		rowWidth := lipgloss.Width(item.Shortcut) + lipgloss.Width(item.Title) + lipgloss.Width(item.Description) + 4
		if rowWidth > innerWidth {
			innerWidth = rowWidth
		}
	}
	return clampDialogWidth(innerWidth+4, 52, 96)
}

func (d *Help) Show()          { d.visible = true }
func (d *Help) Hide()          { d.visible = false }
func (d *Help) Focus() tea.Cmd { return nil }
func (d *Help) Blur()          {}
func (d Help) IsVisible() bool { return d.visible }
