package siftly

import (
	"github.com/charmbracelet/bubbles/key"
)

type Keymap struct {
	Quit                key.Binding
	MarkMode            key.Binding
	ShowMarksOnly       key.Binding
	NextMark            key.Binding
	PrevMark            key.Binding
	Filter              key.Binding
	Search              key.Binding
	ToggleFilter        key.Binding
	SearchNext          key.Binding
	SearchPrev          key.Binding
	CommentOps          key.Binding
	TimeOps             key.Binding
	PageUp              key.Binding
	PageDown            key.Binding
	RowDown             key.Binding
	RowUp               key.Binding
	CommandPalette      key.Binding
	OpenHelp            key.Binding
	ScrollLeft          key.Binding
	ScrollRight         key.Binding
	SaveToFile          key.Binding
	ExportOps           key.Binding
	CopyRow             key.Binding
	Undo                key.Binding
	Redo                key.Binding
	ToggleInspector     key.Binding
	InspectorNextField  key.Binding
	InspectorPrevField  key.Binding
	InspectorCopyField  key.Binding
	InspectorScrollDown key.Binding
	InspectorScrollUp   key.Binding
	SelectRange         key.Binding
	ClearRange          key.Binding
	JumpToStart         key.Binding
	JumpToEnd           key.Binding
	JumpToLineNo        key.Binding
	ToggleGraph         key.Binding
	ReloadFull          key.Binding
	ColumnViewOps       key.Binding
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
	CommentOps: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c e|v", "Comments ops"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "Filter by Regex"),
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
		key.WithKeys("ctrl+u", "pgup"),
		key.WithHelp("ctrl+u/pgup", "Page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("ctrl+d", "pgdown"),
		key.WithHelp("ctrl+d/pgdown", "Page down"),
	),
	RowDown: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "Move a row down"),
	),
	RowUp: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "Move a row up"),
	),
	CommandPalette: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "Command palette"),
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
		key.WithHelp("s", "Save Siftly JSON"),
	),
	ExportOps: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e d", "Export filtered data"),
	),
	CopyRow: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "Copy row/selection"),
	),
	Undo: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "Undo last change"),
	),
	Redo: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "Redo last change"),
	),
	ToggleInspector: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Toggle row inspector"),
	),
	InspectorNextField: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "Next inspector field"),
	),
	InspectorPrevField: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "Previous inspector field"),
	),
	InspectorCopyField: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "Copy inspector field"),
	),
	InspectorScrollDown: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "Scroll inspector down"),
	),
	InspectorScrollUp: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("K", "Scroll inspector up"),
	),
	SelectRange: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "Start/clear row selection"),
	),
	ClearRange: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "Clear row selection"),
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
	TimeOps: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t b|e|r|w", "Time ops"),
	),
	ToggleGraph: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "Toggle graph"),
	),
	ReloadFull: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "Reload full data"),
	),
	ColumnViewOps: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v c|s|o|r", "Columns/Sort/View ops"),
	),
}
