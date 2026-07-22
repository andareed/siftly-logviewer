package dialogs

import (
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

func renderDialogPanel(title string, right string, width int, lines []string, tokenOptions ...ui.DesignTokens) string {
	if width < 6 {
		width = 6
	}

	tokens := dialogDesignTokens(tokenOptions)
	innerWidth := width - 4
	top := renderDialogTopBorder(title, right, width, tokens)
	bottom := tokens.Borders.Strong.Render("└" + strings.Repeat("─", width-2) + "┘")
	flatLines := flattenPanelLines(lines)

	out := make([]string, 0, len(flatLines)+2)
	out = append(out, top)
	for _, line := range flatLines {
		clipped := xansi.Truncate(line, innerWidth, "")
		pad := innerWidth - xansi.StringWidth(clipped)
		if pad < 0 {
			pad = 0
		}
		out = append(out,
			tokens.Borders.Strong.Render("│ ")+
				clipped+strings.Repeat(" ", pad)+
				tokens.Borders.Strong.Render(" │"),
		)
	}
	out = append(out, bottom)
	return strings.Join(out, "\n")
}

func flattenPanelLines(lines []string) []string {
	if len(lines) == 0 {
		return []string{""}
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(strings.ReplaceAll(line, "\r\n", "\n"), "\n")
		out = append(out, parts...)
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}

func renderDialogTopBorder(title string, right string, width int, tokenOptions ...ui.DesignTokens) string {
	tokens := dialogDesignTokens(tokenOptions)
	borderStyle := tokens.Borders.Strong
	title = strings.TrimSpace(title)
	right = strings.TrimSpace(right)
	if title == "" {
		title = "Dialog"
	}

	leftW := lipgloss.Width(title)
	rightW := lipgloss.Width(right)
	filler := width - 2 - leftW - rightW - 4 // spaces around title and right
	if filler < 1 {
		return borderStyle.Render("┌" + strings.Repeat("─", width-2) + "┐")
	}
	if right == "" {
		return borderStyle.Render("┌ ") +
			tokens.Emphasis.Strong.Render(title) +
			borderStyle.Render(" "+strings.Repeat("─", filler+rightW+1)+"┐")
	}
	return borderStyle.Render("┌ ") +
		tokens.Emphasis.Strong.Render(title) +
		borderStyle.Render(" "+strings.Repeat("─", filler)+" ") +
		tokens.Emphasis.Muted.Render(right) +
		borderStyle.Render(" ┐")
}

func dialogSectionLabel(s string, tokenOptions ...ui.DesignTokens) string {
	return dialogDesignTokens(tokenOptions).Emphasis.Muted.Render(s)
}

func dialogStatusLine(kind, msg string, tokenOptions ...ui.DesignTokens) string {
	return dialogDesignTokens(tokenOptions).NoticeStyle(kind).Render(msg)
}

func dialogTopRightState(s string, tokenOptions ...ui.DesignTokens) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return dialogDesignTokens(tokenOptions).Emphasis.Muted.Render(strings.ToUpper(strings.TrimSpace(s)))
}

func renderDialogActionRowWithKeys(innerWidth int, primaryKey, primary string, primaryEnabled bool, secondaryKey, secondary string, tokenOptions ...ui.DesignTokens) string {
	tokens := dialogDesignTokens(tokenOptions)
	primaryLabel := strings.TrimSpace(primary)
	if strings.TrimSpace(primaryKey) != "" {
		primaryLabel = strings.TrimSpace(primaryKey) + " " + primaryLabel
	}
	primaryText := "[ " + primaryLabel + " ]"
	if !primaryEnabled {
		primaryText = tokens.Emphasis.Subtle.Render(primaryText)
	} else {
		primaryText = tokens.States.Accent.Render(primaryText)
	}

	secondaryText := strings.TrimSpace(secondary)
	if strings.TrimSpace(secondaryKey) != "" && secondaryText != "" {
		secondaryText = strings.TrimSpace(secondaryKey) + " " + secondaryText
	}
	if secondaryText != "" {
		secondaryText = tokens.Emphasis.Muted.Render(secondaryText)
	}

	row := primaryText
	if secondaryText != "" {
		row += "   " + secondaryText
	}

	pad := innerWidth - xansi.StringWidth(row)
	if pad <= 0 {
		return row
	}
	return strings.Repeat(" ", pad) + row
}

func dialogDesignTokens(options []ui.DesignTokens) ui.DesignTokens {
	if len(options) > 0 {
		return ui.ResolveDesignTokens(options[0])
	}
	return ui.DefaultDesignTokens()
}

func styleDialogTextInput(input *textinput.Model, tokens ui.DesignTokens) {
	if input == nil {
		return
	}
	tokens = ui.ResolveDesignTokens(tokens)
	input.PromptStyle = tokens.Emphasis.Strong
	input.TextStyle = tokens.Emphasis.Normal
	input.PlaceholderStyle = tokens.Emphasis.Subtle
	input.CompletionStyle = tokens.Emphasis.Muted
	input.Cursor.Style = tokens.States.Accent
}
