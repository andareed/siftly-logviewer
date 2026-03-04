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
)

type Action struct {
	Kind    ActionKind
	Path    string
	Pattern string
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
