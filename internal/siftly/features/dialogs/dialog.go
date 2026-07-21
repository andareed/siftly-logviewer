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
)

type Action struct {
	Kind      ActionKind
	Path      string
	Pattern   string
	CommandID string
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
