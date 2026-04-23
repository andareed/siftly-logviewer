package siftly

import (
	"github.com/charmbracelet/bubbles/key"
)

type Keymap struct {
	Quit          key.Binding
	MarkMode      key.Binding
	ShowMarksOnly key.Binding
	NextMark      key.Binding
	PrevMark      key.Binding
	Filter        key.Binding
	Sort          key.Binding
	Search        key.Binding
	ToggleFilter  key.Binding
	SearchNext    key.Binding
	SearchPrev    key.Binding
	TimeWindowSet key.Binding
	ShowComment   key.Binding
	EditComment   key.Binding
	CommentOps    key.Binding
	TimeOps       key.Binding
	PageUp        key.Binding
	PageDown      key.Binding
	RowDown       key.Binding
	RowUp         key.Binding
	OpenHelp      key.Binding
	ScrollLeft    key.Binding
	ScrollRight   key.Binding
	SaveToFile    key.Binding
	ExportToFile  key.Binding
	CopyRow       key.Binding
	JumpToStart   key.Binding
	JumpToEnd     key.Binding
	JumpToLineNo  key.Binding
	TimeWindow    key.Binding
	ToggleGraph   key.Binding
	ColumnViewOps key.Binding
}

var Keys = Keymap{
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "Quit"),
	),
	MarkMode: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "Mark mode"),
	),
	ShowMarksOnly: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "Toggle show only marked"),
	),
	NextMark: key.NewBinding(
		key.WithKeys("]", "ctrl+n"),
		key.WithHelp("]/ctrl+n", "Jump to next mark"),
	),
	PrevMark: key.NewBinding(
		key.WithKeys("[", "ctrl+p"),
		key.WithHelp("[/ctrl+p", "Jump to previous mark"),
	),
	ShowComment: key.NewBinding(
		key.WithKeys("V"),
		key.WithHelp("V", "View comments"),
	),
	EditComment: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "Edit comment on selected row"),
	),
	CommentOps: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c e|v", "Comments ops"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "Filter by Regex"),
	),
	Sort: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "Sort rows"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "Search"),
	),
	ToggleFilter: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "Toggle current filter"),
	),
	SearchNext: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "Next search"),
	),
	SearchPrev: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "Prev search"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("u", "pgup"),
		key.WithHelp("u/pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("d", "pgdown"),
		key.WithHelp("d/pgdown", "Page down"),
	),
	RowDown: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "Move a row down"),
	),
	RowUp: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "Move a row up"),
	),
	OpenHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "Help / keys"),
	),
	ScrollLeft: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h or <- ", "Scroll the grid left"),
	),
	ScrollRight: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l or >- ", "Scroll the grid right"),
	),
	SaveToFile: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "Save to filename"),
	),
	ExportToFile: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "Export to filename"),
	),
	CopyRow: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "Copy row to clipboard"),
	),
	JumpToStart: key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g/home", "Jump to start"),
	),
	JumpToEnd: key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G/end", "Jump to end"),
	),
	JumpToLineNo: key.NewBinding(
		key.WithKeys(":"),
		key.WithHelp(":", "Jump To line number"),
	),
	TimeWindow: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "Time window"),
	),
	TimeOps: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t b|e|r|w", "Time ops"),
	),
	TimeWindowSet: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "Set time window from cursor"),
	),
	ToggleGraph: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "Toggle graph"),
	),
	ColumnViewOps: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v c|s|o|r", "Columns/Sort/View ops"),
	),
}

func (k Keymap) Legend(graphEnabled bool) []key.Binding {
	legend := []key.Binding{
		k.Quit,
		k.MarkMode,
		k.ShowMarksOnly,
		k.NextMark,
		k.PrevMark,
		k.Filter,
		k.Search,
		k.ToggleFilter,
		k.SearchNext,
		k.SearchPrev,
		k.TimeOps,
		k.CommentOps,
		k.PageUp,
		k.PageDown,
		k.CopyRow,
		k.ExportToFile,
		k.SaveToFile,
		k.RowUp,
		k.RowDown,
		k.JumpToStart,
		k.JumpToEnd,
		k.JumpToLineNo,
		k.TimeWindow,
		k.ColumnViewOps,
	}
	if graphEnabled {
		legend = append(legend, k.ToggleGraph)
	}
	return legend
}
