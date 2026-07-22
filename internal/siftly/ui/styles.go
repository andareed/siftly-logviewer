package ui

import "github.com/charmbracelet/lipgloss"

// Styles defines the UI look-and-feel that a wrapper package injects.
type Styles struct {
	Tokens             DesignTokens
	App                lipgloss.Style
	Header             lipgloss.Style
	Row                lipgloss.Style
	RowSelected        lipgloss.Style
	RepeatedCell       lipgloss.Style
	RowText            lipgloss.Style
	RowSelectedText    lipgloss.Style
	Cell               lipgloss.Style
	Input              lipgloss.Style
	Table              lipgloss.Style
	GraphArea          lipgloss.Style
	RedMarker          lipgloss.Style
	GreenMarker        lipgloss.Style
	AmberMarker        lipgloss.Style
	CommentArea        lipgloss.Style
	TimeWindowArea     lipgloss.Style
	SearchHighlight    lipgloss.Style
	RowTextFGColor     lipgloss.Color
	RowSelectedFG      lipgloss.Color
	RowSelectedBG      lipgloss.Color
	RowRangeSelectedBG lipgloss.Color
	DefaultMarker      string
	PillMarker         string
	CommentMarker      string
}

// DefaultStyles builds all component styles from the shared design tokens.
func DefaultStyles() Styles {
	tokens := DefaultDesignTokens()
	colors := tokens.Colors

	table := lipgloss.NewStyle().
		BorderStyle(tokens.Borders.Shape).
		BorderForeground(colors.Border)
	inset := lipgloss.NewStyle().
		Border(tokens.Borders.Shape).
		BorderForeground(colors.BorderStrong).
		Padding(0, 0).
		BorderLeft(true)

	return Styles{
		Tokens:          tokens,
		App:             lipgloss.NewStyle().Margin(1, 2),
		Header:          tokens.Emphasis.Strong.Background(colors.SurfaceHeader),
		Row:             tokens.Emphasis.Normal,
		RowSelected:     tokens.States.Selected,
		RepeatedCell:    tokens.Emphasis.Subtle,
		RowText:         tokens.Emphasis.Normal,
		RowSelectedText: lipgloss.NewStyle().Foreground(colors.TextSelected),
		Cell:            lipgloss.NewStyle().Padding(0, 1),
		Input: lipgloss.NewStyle().
			Border(tokens.Borders.Shape, true).
			BorderForeground(colors.BorderStrong).
			Padding(1),
		Table:              table,
		GraphArea:          table,
		RedMarker:          lipgloss.NewStyle().Foreground(colors.MarkRed),
		GreenMarker:        lipgloss.NewStyle().Foreground(colors.MarkGreen),
		AmberMarker:        lipgloss.NewStyle().Foreground(colors.MarkAmber),
		CommentArea:        inset,
		TimeWindowArea:     inset,
		SearchHighlight:    tokens.States.SearchMatch,
		RowTextFGColor:     colors.Text,
		RowSelectedFG:      colors.TextSelected,
		RowSelectedBG:      colors.SurfaceSelected,
		RowRangeSelectedBG: colors.SurfaceRange,
		DefaultMarker:      " ",
		PillMarker:         "▐",
		CommentMarker:      "[*]",
	}
}

// ResolvedTokens returns the injected token set or the default for partial
// style bundles used by tests and embedders.
func (styles Styles) ResolvedTokens() DesignTokens {
	return ResolveDesignTokens(styles.Tokens)
}
