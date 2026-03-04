package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type FooterState struct {
	ModeLabel string
	Prompt    string

	StatusMessage string
	Hints         string
	IsInputMode   bool
}

type FooterStyles struct {
	BarBG      lipgloss.Color
	StatusBG   lipgloss.Color
	ModePillBG lipgloss.Color
	ModePillFG lipgloss.Color
	FileNameFG lipgloss.Color
	TextFG     lipgloss.Color
	DimFG      lipgloss.Color
	StatusFG   lipgloss.Color
	LegendFG   lipgloss.Color
}

func DefaultFooterStyles() FooterStyles {
	return FooterStyles{
		BarBG:      lipgloss.Color("#2b2b2b"),
		StatusBG:   lipgloss.Color("#000000"),
		ModePillBG: lipgloss.Color("#ff9f1c"),
		ModePillFG: lipgloss.Color("#000000"),
		FileNameFG: lipgloss.Color("#e0e0e0"),
		TextFG:     lipgloss.Color("#cfcfcf"),
		DimFG:      lipgloss.Color("#a0a0a0"),
		StatusFG:   lipgloss.Color("#9a9a9a"),
		LegendFG:   lipgloss.Color("#b0b0b0"),
	}
}

func RenderFooter(width int, st FooterState, styles FooterStyles) string {
	if width <= 0 {
		return ""
	}
	st = normalizeFooterState(st)

	modeLine := renderModeLine(width, st, styles)
	promptLine := renderPromptLine(width, st, styles)
	statusLine := renderStatusLine(width, st, styles)
	return modeLine + "\n" + promptLine + "\n" + statusLine
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

func renderModeLine(width int, st FooterState, styles FooterStyles) string {
	mode := strings.ToUpper(strings.TrimSpace(st.ModeLabel))
	if !strings.HasSuffix(mode, " MODE") {
		mode += " MODE"
	}

	text := padRightRunes(truncateRunes(mode, width), width)
	lineStyle := lipgloss.NewStyle().Foreground(styles.TextFG).Background(styles.BarBG)
	if st.IsInputMode {
		lineStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.ModePillFG).Background(styles.ModePillBG)
	}
	return lineStyle.Render(text)
}

func renderPromptLine(width int, st FooterState, styles FooterStyles) string {
	line := ""
	if st.IsInputMode {
		line = "> " + clipInputKeepTail(st.Prompt, width-2)
	} else {
		prompt := strings.TrimSpace(st.Prompt)
		if prompt != "" {
			line = "> " + clipInputKeepTail(prompt, width-2)
		}
	}

	line = padRightRunes(truncateRunes(line, width), width)
	style := lipgloss.NewStyle().Background(styles.StatusBG).Foreground(styles.TextFG)
	return style.Render(line)
}

func renderStatusLine(width int, st FooterState, styles FooterStyles) string {
	line := padRightRunes(truncateRunes(strings.TrimSpace(st.StatusMessage), width), width)
	style := lipgloss.NewStyle().Background(styles.StatusBG).Foreground(styles.StatusFG)
	return style.Render(line)
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
