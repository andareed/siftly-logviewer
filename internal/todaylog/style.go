package todaylog

import (
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
)

const (
	rowTextFGColor         = "#c0c0c0"
	rowSelectedTextFGColor = "#e0e0e0"
	rowSelectedBGColor     = "#3a3a3a"
)

var (
	headerStyle = lipgloss.NewStyle().BorderStyle(lipgloss.Border{
		Left:  " ",
		Right: " ",
	}).BorderLeft(true).BorderRight(true)
	cellStyle  = lipgloss.NewStyle().Padding(0, 1)
	appStyle   = lipgloss.NewStyle().Margin(1, 2)
	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	graphStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(rowTextFGColor))
	rowSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(rowSelectedBGColor)).
				Foreground(lipgloss.Color(rowSelectedTextFGColor))
	redMarker      = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	greenMarker    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	amberMarker    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	commentArea    = tableStyle
	timeWindowArea = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("245")).
			Padding(0, 0).
			BorderLeft(true)
	searchHighlight = lipgloss.NewStyle().
			Background(lipgloss.Color("#f5c542")).
			Foreground(lipgloss.Color("#000000"))
)

func SiftlyStyles() ui.Styles {
	return ui.Styles{
		App:             appStyle,
		Header:          headerStyle,
		Row:             rowStyle,
		RowSelected:     rowSelectedStyle,
		Cell:            cellStyle,
		Table:           tableStyle,
		GraphArea:       graphStyle,
		RedMarker:       redMarker,
		GreenMarker:     greenMarker,
		AmberMarker:     amberMarker,
		CommentArea:     commentArea,
		TimeWindowArea:  timeWindowArea,
		SearchHighlight: searchHighlight,
		RowTextFGColor:  lipgloss.Color(rowTextFGColor),
		RowSelectedFG:   lipgloss.Color(rowSelectedTextFGColor),
		RowSelectedBG:   lipgloss.Color(rowSelectedBGColor),
		DefaultMarker:   " ",
		PillMarker:      "▐",
		CommentMarker:   "[*]",
	}
}
