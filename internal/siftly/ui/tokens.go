package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorTokens names colours by purpose so components do not maintain their own
// palettes.
type ColorTokens struct {
	SurfaceFooter      lipgloss.Color
	SurfaceFooterInset lipgloss.Color
	SurfaceHeader      lipgloss.Color
	SurfaceSelected    lipgloss.Color
	SurfaceRange       lipgloss.Color
	SurfaceOverlay     lipgloss.Color
	Text               lipgloss.Color
	TextStrong         lipgloss.Color
	TextSelected       lipgloss.Color
	TextMuted          lipgloss.Color
	TextSubtle         lipgloss.Color
	TextInverse        lipgloss.Color
	Border             lipgloss.Color
	BorderStrong       lipgloss.Color
	Accent             lipgloss.Color
	Success            lipgloss.Color
	Warning            lipgloss.Color
	Error              lipgloss.Color
	SearchBackground   lipgloss.Color
	SearchForeground   lipgloss.Color
	MarkRed            lipgloss.Color
	MarkGreen          lipgloss.Color
	MarkAmber          lipgloss.Color
}

// EmphasisTokens provides the four text levels used throughout the TUI.
type EmphasisTokens struct {
	Strong lipgloss.Style
	Normal lipgloss.Style
	Muted  lipgloss.Style
	Subtle lipgloss.Style
}

// BorderTokens centralises frame shape and visual weight.
type BorderTokens struct {
	Shape  lipgloss.Border
	Subtle lipgloss.Style
	Strong lipgloss.Style
}

// StateTokens defines consistent interaction and notice states.
type StateTokens struct {
	Selected    lipgloss.Style
	Range       lipgloss.Style
	SearchMatch lipgloss.Style
	Accent      lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Error       lipgloss.Style
}

// DesignTokens is the shared visual vocabulary for tables, panels, dialogs,
// notices and graphs.
type DesignTokens struct {
	Colors      ColorTokens
	Emphasis    EmphasisTokens
	Borders     BorderTokens
	States      StateTokens
	GraphSeries []lipgloss.Color
}

// DefaultDesignTokens returns a fresh copy of Siftly's default theme.
func DefaultDesignTokens() DesignTokens {
	colors := ColorTokens{
		SurfaceFooter:      lipgloss.Color("#202020"),
		SurfaceFooterInset: lipgloss.Color("#101010"),
		SurfaceHeader:      lipgloss.Color("#303030"),
		SurfaceSelected:    lipgloss.Color("#3a3a3a"),
		SurfaceRange:       lipgloss.Color("#2a2a2a"),
		SurfaceOverlay:     lipgloss.Color("#303030"),
		Text:               lipgloss.Color("#c0c0c0"),
		TextStrong:         lipgloss.Color("#f0f0f0"),
		TextSelected:       lipgloss.Color("#e0e0e0"),
		TextMuted:          lipgloss.Color("#a0a0a0"),
		TextSubtle:         lipgloss.Color("#777777"),
		TextInverse:        lipgloss.Color("#000000"),
		Border:             lipgloss.Color("#585858"),
		BorderStrong:       lipgloss.Color("#8a8a8a"),
		Accent:             lipgloss.Color("#ff9f1c"),
		Success:            lipgloss.Color("#72d99c"),
		Warning:            lipgloss.Color("#f0b45a"),
		Error:              lipgloss.Color("#ff6b6b"),
		SearchBackground:   lipgloss.Color("#f5c542"),
		SearchForeground:   lipgloss.Color("#000000"),
		MarkRed:            lipgloss.Color("1"),
		MarkGreen:          lipgloss.Color("2"),
		MarkAmber:          lipgloss.Color("3"),
	}

	return DesignTokens{
		Colors: colors,
		Emphasis: EmphasisTokens{
			Strong: lipgloss.NewStyle().Bold(true).Foreground(colors.TextStrong),
			Normal: lipgloss.NewStyle().Foreground(colors.Text),
			Muted:  lipgloss.NewStyle().Foreground(colors.TextMuted),
			Subtle: lipgloss.NewStyle().Foreground(colors.TextSubtle),
		},
		Borders: BorderTokens{
			Shape:  lipgloss.NormalBorder(),
			Subtle: lipgloss.NewStyle().Foreground(colors.Border),
			Strong: lipgloss.NewStyle().Foreground(colors.BorderStrong),
		},
		States: StateTokens{
			Selected: lipgloss.NewStyle().
				Foreground(colors.TextSelected).
				Background(colors.SurfaceSelected),
			Range: lipgloss.NewStyle().
				Foreground(colors.Text).
				Background(colors.SurfaceRange),
			SearchMatch: lipgloss.NewStyle().
				Foreground(colors.SearchForeground).
				Background(colors.SearchBackground),
			Accent:  lipgloss.NewStyle().Bold(true).Foreground(colors.Accent),
			Success: lipgloss.NewStyle().Foreground(colors.Success),
			Warning: lipgloss.NewStyle().Foreground(colors.Warning),
			Error:   lipgloss.NewStyle().Foreground(colors.Error),
		},
		GraphSeries: []lipgloss.Color{
			lipgloss.Color("#ff6b6b"),
			lipgloss.Color("#4ecdc4"),
			lipgloss.Color("#ffe66d"),
			lipgloss.Color("#5b8def"),
			lipgloss.Color("#a66dd4"),
			lipgloss.Color("#f08a24"),
		},
	}
}

// ResolveDesignTokens supplies the default theme for zero-value style bundles.
func ResolveDesignTokens(tokens DesignTokens) DesignTokens {
	if tokens.Colors.Text == "" {
		return DefaultDesignTokens()
	}
	return tokens
}

// NoticeStyle maps a semantic notice kind to its shared state style.
func (tokens DesignTokens) NoticeStyle(kind string) lipgloss.Style {
	tokens = ResolveDesignTokens(tokens)
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "success":
		return tokens.States.Success
	case "warn":
		return tokens.States.Warning
	case "error":
		return tokens.States.Error
	default:
		return tokens.Emphasis.Muted
	}
}
