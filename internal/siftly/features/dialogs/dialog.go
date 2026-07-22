package dialogs

import tea "github.com/charmbracelet/bubbletea"

type ActionKind int

const (
	ActionNone ActionKind = iota
	ActionClose
	ActionSaveConfirm
	ActionSaveCancel
	ActionExportConfirm
	ActionExportCancel
	ActionFilterApply
	ActionFilterCancel
	ActionCommandRun
	ActionCommandCancel
	ActionColumnManagerApply
	ActionColumnManagerCancel
)

type Action struct {
	Kind          ActionKind
	Path          string
	Pattern       string
	CommandID     string
	ColumnManager *ColumnManagerResult
}

type CommandItem struct {
	ID             string
	Category       string
	Title          string
	Shortcut       string
	Description    string
	Keywords       string
	Enabled        bool
	DisabledReason string
}

type ColumnManagerItem struct {
	SourceIndex int
	Name        string
	Visible     bool
	Frozen      bool
	MinWidth    int
	Weight      float64
	AutoFit     bool
}

type ColumnManagerResult struct {
	Columns     []ColumnManagerItem
	SortEnabled bool
	SortColumn  int
	SortDesc    bool
}

// Dialog is the common interface all dialogs (Save, Export, Help, etc.) implement.
// It keeps your model logic generic.
type Dialog interface {
	Init() tea.Cmd // optional, can return nil
	Update(msg tea.Msg) (Dialog, Action, tea.Cmd)
	View() string

	Focus() tea.Cmd
	Blur()
	IsVisible() bool
	Show()
	Hide()
}

// Resizable is implemented by dialogs that adapt while they are open.
type Resizable interface {
	Resize(terminalWidth, terminalHeight int)
}
