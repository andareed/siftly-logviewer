package dialogs

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

func renderDialogPanel(title string, right string, width int, lines []string) string {
	if width < 6 {
		width = 6
	}

	innerWidth := width - 4
	top := renderDialogTopBorder(title, right, width)
	bottom := "└" + strings.Repeat("─", width-2) + "┘"
	flatLines := flattenPanelLines(lines)

	out := make([]string, 0, len(flatLines)+2)
	out = append(out, top)
	for _, line := range flatLines {
		clipped := xansi.Truncate(line, innerWidth, "")
		pad := innerWidth - xansi.StringWidth(clipped)
		if pad < 0 {
			pad = 0
		}
		out = append(out, "│ "+clipped+strings.Repeat(" ", pad)+" │")
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

func renderDialogTopBorder(title string, right string, width int) string {
	title = strings.TrimSpace(title)
	right = strings.TrimSpace(right)
	if title == "" {
		title = "Dialog"
	}

	leftW := lipgloss.Width(title)
	rightW := lipgloss.Width(right)
	filler := width - 2 - leftW - rightW - 4 // spaces around title and right
	if filler < 1 {
		return "┌" + strings.Repeat("─", width-2) + "┐"
	}
	if right == "" {
		return "┌ " + title + " " + strings.Repeat("─", filler+rightW+1) + "┐"
	}
	return "┌ " + title + " " + strings.Repeat("─", filler) + " " + right + " ┐"
}

func dialogSectionLabel(s string) string {
	return lipgloss.NewStyle().Faint(true).Render(s)
}

func dialogStatusLine(kind, msg string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "success":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(msg)
	case "warn":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(msg)
	case "error":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(msg)
	default:
		return msg
	}
}

func dialogTopRightState(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return lipgloss.NewStyle().Faint(true).Render(strings.ToUpper(strings.TrimSpace(s)))
}

func renderDialogActionRowWithKeys(innerWidth int, primaryKey, primary string, primaryEnabled bool, secondaryKey, secondary string) string {
	primaryLabel := strings.TrimSpace(primary)
	if strings.TrimSpace(primaryKey) != "" {
		primaryLabel = strings.TrimSpace(primaryKey) + " " + primaryLabel
	}
	primaryText := "[ " + primaryLabel + " ]"
	if !primaryEnabled {
		primaryText = lipgloss.NewStyle().Faint(true).Render(primaryText)
	} else {
		primaryText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).Render(primaryText)
	}

	secondaryText := strings.TrimSpace(secondary)
	if strings.TrimSpace(secondaryKey) != "" && secondaryText != "" {
		secondaryText = strings.TrimSpace(secondaryKey) + " " + secondaryText
	}
	if secondaryText != "" {
		secondaryText = lipgloss.NewStyle().Faint(true).Render(secondaryText)
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
