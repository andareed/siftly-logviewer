package ui

import "github.com/charmbracelet/lipgloss"

// Styles defines the UI look-and-feel that a wrapper package injects.
type Styles struct {
	App             lipgloss.Style
	Header          lipgloss.Style
	Row             lipgloss.Style
	RowSelected     lipgloss.Style
	RowText         lipgloss.Style
	RowSelectedText lipgloss.Style
	Cell            lipgloss.Style
	Input           lipgloss.Style
	Table           lipgloss.Style
	GraphArea       lipgloss.Style
	RedMarker       lipgloss.Style
	GreenMarker     lipgloss.Style
	AmberMarker     lipgloss.Style
	CommentArea     lipgloss.Style
	TimeWindowArea  lipgloss.Style
	SearchHighlight lipgloss.Style
	RowTextFGColor  lipgloss.Color
	RowSelectedFG   lipgloss.Color
	RowSelectedBG   lipgloss.Color
	DefaultMarker   string
	PillMarker      string
	CommentMarker   string
}
