package devfmt

import (
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
)

const (
	rowTextFGColor         = "#c0c0c0"
	rowSelectedTextFGColor = "#e0e0e0"
	rowSelectedBGColor     = "#3a3a3a"
	searchHighlightBGColor = "#f5c542"
	searchHighlightFGColor = "#000000"
)

var (
	appstyle    = lipgloss.NewStyle().Margin(1, 2)
	headerStyle = lipgloss.NewStyle().BorderStyle(lipgloss.Border{
		Left:  " ",
		Right: " ",
	}).BorderLeft(true).BorderRight(true)
	rowStyle         = lipgloss.NewStyle()
	rowSelectedStyle = lipgloss.NewStyle().Background(lipgloss.Color(rowSelectedBGColor))

	rowTextStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color(rowTextFGColor))
	rowSelectedTextstyle = lipgloss.NewStyle().Foreground(lipgloss.Color(rowSelectedTextFGColor))

	cellStyle     = lipgloss.NewStyle().Padding(0, 1)
	inputStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Padding(1)
	tableStyle    = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	redMarker     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	greenMarker   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	amberMarker   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	defaultMarker = " "
	pillMarker    = "▐"
	commentMarker = "[*]"

	commentArea = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("245")).
			Padding(0, 0).BorderLeft(true)

	timeWindowArea = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("245")).
			Padding(0, 0).BorderLeft(true)

	searchHighlight = lipgloss.NewStyle().
			Background(lipgloss.Color(searchHighlightBGColor)).
			Foreground(lipgloss.Color(searchHighlightFGColor))
)

func SiftlyStyles() ui.Styles {
	return ui.Styles{
		App:             appstyle,
		Header:          headerStyle,
		Row:             rowStyle,
		RowSelected:     rowSelectedStyle,
		RowText:         rowTextStyle,
		RowSelectedText: rowSelectedTextstyle,
		Cell:            cellStyle,
		Input:           inputStyle,
		Table:           tableStyle,
		GraphArea:       tableStyle,
		RedMarker:       redMarker,
		GreenMarker:     greenMarker,
		AmberMarker:     amberMarker,
		CommentArea:     commentArea,
		TimeWindowArea:  timeWindowArea,
		SearchHighlight: searchHighlight,
		RowTextFGColor:  lipgloss.Color(rowTextFGColor),
		RowSelectedFG:   lipgloss.Color(rowSelectedTextFGColor),
		RowSelectedBG:   lipgloss.Color(rowSelectedBGColor),
		DefaultMarker:   defaultMarker,
		PillMarker:      pillMarker,
		CommentMarker:   commentMarker,
	}
}
