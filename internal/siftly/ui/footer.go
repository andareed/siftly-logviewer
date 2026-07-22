package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

type FooterState struct {
	ModeLabel    string
	ActiveStates []string
	Prompt       string

	StatusMessage string
	StatusKind    string
	Hints         string
	IsInputMode   bool
}

type FooterStyles struct {
	BarBG      lipgloss.Color
	StatusBG   lipgloss.Color
	ModeBG     lipgloss.Color
	ModeFG     lipgloss.Color
	ModePillBG lipgloss.Color
	ModePillFG lipgloss.Color
	FileNameFG lipgloss.Color
	TextFG     lipgloss.Color
	DimFG      lipgloss.Color
	StatusFG   lipgloss.Color
	SuccessFG  lipgloss.Color
	WarnFG     lipgloss.Color
	ErrorFG    lipgloss.Color
	LegendFG   lipgloss.Color
}

func DefaultFooterStyles() FooterStyles {
	return FooterStylesFromTokens(DefaultDesignTokens())
}

// FooterStylesFromTokens maps the shared semantic palette onto the footer's
// two-line layout.
func FooterStylesFromTokens(tokens DesignTokens) FooterStyles {
	tokens = ResolveDesignTokens(tokens)
	colors := tokens.Colors
	return FooterStyles{
		BarBG:      colors.SurfaceFooter,
		StatusBG:   colors.SurfaceFooterInset,
		ModeBG:     colors.SurfaceSelected,
		ModeFG:     colors.TextStrong,
		ModePillBG: colors.Accent,
		ModePillFG: colors.TextInverse,
		FileNameFG: colors.TextSelected,
		TextFG:     colors.Text,
		DimFG:      colors.TextMuted,
		StatusFG:   colors.TextMuted,
		SuccessFG:  colors.Success,
		WarnFG:     colors.Warning,
		ErrorFG:    colors.Error,
		LegendFG:   colors.TextMuted,
	}
}

func RenderFooter(width int, st FooterState, styles FooterStyles) string {
	if width <= 0 {
		return ""
	}
	st = normalizeFooterState(st)

	stateLine := renderStateLine(width, st, styles)
	actionLine := renderActionLine(width, st, styles)
	return stateLine + "\n" + actionLine
}

func normalizeFooterState(st FooterState) FooterState {
	if strings.TrimSpace(st.ModeLabel) == "" {
		st.ModeLabel = "NORMAL"
	}
	if st.Hints == "" {
		st.Hints = "v view · c comment · t time · / search · ? help"
	}
	return st
}

func renderStateLine(width int, st FooterState, styles FooterStyles) string {
	mode := strings.ToUpper(strings.TrimSpace(st.ModeLabel))
	modeStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.ModeFG).Background(styles.ModeBG)
	if st.IsInputMode {
		modeStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.ModePillFG).Background(styles.ModePillBG)
	}
	left := modeStyle.Render(" " + mode + " ")

	stateStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.TextFG).Background(styles.BarBG)
	warnStateStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.WarnFG).Background(styles.BarBG)
	separatorStyle := lipgloss.NewStyle().Foreground(styles.DimFG).Background(styles.BarBG)
	for _, state := range st.ActiveStates {
		state = strings.TrimSpace(state)
		if state == "" {
			continue
		}
		style := stateStyle
		if strings.EqualFold(state, "UNSAVED") {
			style = warnStateStyle
		}
		left += separatorStyle.Render(" │ ") + style.Render(state)
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor(st.StatusKind, styles)).Background(styles.BarBG)
	right := ""
	if status := strings.TrimSpace(st.StatusMessage); status != "" {
		right = statusStyle.Render(status)
	}

	if right == "" {
		return fillFooterLine(xansi.Truncate(left, width, ""), width, styles.BarBG)
	}
	if xansi.StringWidth(right) > width/2 {
		right = xansi.Truncate(right, maxInt(1, width/2), "…")
	}
	leftWidth := width - xansi.StringWidth(right) - 1
	if leftWidth < 1 {
		return fillFooterLine(xansi.Truncate(right, width, ""), width, styles.BarBG)
	}
	left = xansi.Truncate(left, leftWidth, "")
	gap := width - xansi.StringWidth(left) - xansi.StringWidth(right)
	return left + lipgloss.NewStyle().Background(styles.BarBG).Render(strings.Repeat(" ", maxInt(0, gap))) + right
}

func renderActionLine(width int, st FooterState, styles FooterStyles) string {
	background := lipgloss.NewStyle().Background(styles.StatusBG)
	hint := strings.TrimSpace(st.Hints)
	if !st.IsInputMode {
		line := lipgloss.NewStyle().Foreground(styles.DimFG).Background(styles.StatusBG).
			Render(fitFooterHints(hint, width))
		return fillFooterLine(line, width, styles.StatusBG)
	}

	hintWidth := minInt(lipgloss.Width(hint), width/2)
	promptWidth := width
	if hintWidth > 0 {
		promptWidth = width - hintWidth - 2
		if promptWidth < 12 {
			hintWidth = 0
			promptWidth = width
		}
	}
	prompt := "> " + clipInputKeepTail(st.Prompt, maxInt(0, promptWidth-2))
	prompt = lipgloss.NewStyle().Foreground(styles.TextFG).Background(styles.StatusBG).Render(prompt)
	if hintWidth == 0 {
		return fillFooterLine(xansi.Truncate(prompt, width, ""), width, styles.StatusBG)
	}
	hints := fitFooterHints(hint, hintWidth)
	hints = lipgloss.NewStyle().Foreground(styles.DimFG).Background(styles.StatusBG).Render(hints)
	gap := width - xansi.StringWidth(prompt) - xansi.StringWidth(hints)
	return prompt + background.Render(strings.Repeat(" ", maxInt(0, gap))) + hints
}

func fitFooterHints(hints string, width int) string {
	hints = strings.TrimSpace(hints)
	if width <= 0 || hints == "" {
		return ""
	}
	if lipgloss.Width(hints) <= width {
		return hints
	}
	parts := strings.Split(hints, "   ")
	fitted := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		candidate := part
		if fitted != "" {
			candidate = fitted + "   " + part
		}
		if lipgloss.Width(candidate) > width {
			break
		}
		fitted = candidate
	}
	if fitted != "" {
		return fitted
	}
	return truncateRunes(hints, width)
}

func statusColor(kind string, styles FooterStyles) lipgloss.Color {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "success":
		return styles.SuccessFG
	case "warn":
		return styles.WarnFG
	case "error":
		return styles.ErrorFG
	default:
		return styles.StatusFG
	}
}

func fillFooterLine(line string, width int, background lipgloss.Color) string {
	line = xansi.Truncate(line, width, "")
	padding := width - xansi.StringWidth(line)
	if padding <= 0 {
		return line
	}
	return line + lipgloss.NewStyle().Background(background).Render(strings.Repeat(" ", padding))
}

func clipInputKeepTail(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxW {
		return s
	}
	if maxW <= 1 {
		return string(r[len(r)-maxW:])
	}
	return "…" + string(r[len(r)-(maxW-1):])
}

func truncateRunes(s string, w int) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	return string(r[:w])
}

func padRightRunes(s string, w int) string {
	cur := lipgloss.Width(s)
	if cur >= w {
		return s
	}
	return s + strings.Repeat(" ", w-cur)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
