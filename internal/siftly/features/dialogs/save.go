package dialogs

import (
	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Save dialog (modal) ----------------------------------------------------

type Save struct {
	input          textinput.Model
	visible        bool
	lastDir        string
	targetPath     string
	targetDir      string
	fileLines      []string
	state          fileDialogState
	tokens         ui.DesignTokens
	width          int
	height         int
	terminalWidth  int
	terminalHeight int
}

func (d Save) Init() tea.Cmd { return d.input.Focus() }

func NewSaveDialog(defaultName, lastDir string, tokenOptions ...ui.DesignTokens) *Save {
	ti := textinput.New()
	// Prompt and placeholder
	ti.Placeholder = defaultName
	ti.Prompt = ""
	ti.CharLimit = 256
	// Wide enough for typical paths
	ti.Width = 50
	if defaultName != "" {
		ti.SetValue(defaultName)
	}
	tokens := dialogDesignTokens(tokenOptions)
	styleDialogTextInput(&ti, tokens)
	d := &Save{input: ti, visible: true, lastDir: lastDir, tokens: tokens}
	d.refreshPreview()
	d.Resize(82, 29)
	return d
}

func (d *Save) Update(msg tea.Msg) (Dialog, Action, tea.Cmd) {
	logging.Debug("SaveDialog:Update:: Called")
	if !d.visible {
		return d, Action{Kind: ActionNone}, nil
	}
	switch m := msg.(type) {
	case tea.KeyMsg:
		logging.Debug("Update::Handle Key Message")
		s := m.String()
		switch s {
		case "enter":
			logging.Infof("SaveDialog:Update::Enter key was pressed, starting saving to file.")
			d.refreshPreview()
			if !d.state.PrimaryEnabled || d.targetPath == "" {
				return d, Action{Kind: ActionNone}, nil
			}
			return d, Action{Kind: ActionSaveConfirm, Path: d.targetPath}, nil
		case "esc":
			logging.Debug("SaveDialog:Update::Esc key was prssed, cancel anything to do with this")
			return d, Action{Kind: ActionSaveCancel}, nil
		}
	}
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	d.refreshPreview()
	return d, Action{Kind: ActionNone}, cmd
}

func (d Save) View() string {
	if !d.visible {
		return ""
	}
	innerWidth := max(1, d.width-4)
	contentLines := fileDialogContent(d.input.View(), d.state, innerWidth, d.height, d.tokens)
	return renderDialogPanel("Save Siftly JSON", dialogTopRightState(d.state.TopRightState, d.tokens), d.width, contentLines, d.tokens)
}

func (d *Save) Show() {
	d.visible = true
	d.input.Focus()
}

func (d *Save) Hide() {
	d.visible = false
	d.input.Blur()
}

func (d *Save) Focus() tea.Cmd { return d.input.Focus() }
func (d *Save) Blur()          { d.input.Blur() }
func (d Save) IsVisible() bool { return d.visible }

func (d *Save) Resize(terminalWidth, terminalHeight int) {
	d.terminalWidth = terminalWidth
	d.terminalHeight = terminalHeight
	d.width = responsiveDialogWidth(terminalWidth, preferredFileDialogWidth("Save Siftly JSON", d.input.Value(), d.state), 44)
	preferredHeight := min(25, 12+len(d.fileLines))
	d.height = responsiveDialogHeight(terminalHeight, preferredHeight)
	d.input.Width = max(1, d.width-8)
}

func (d *Save) refreshPreview() {
	d.state = resolveFileDialogState(d.lastDir, d.input.Value(), d.input.Placeholder, "Save")
	d.targetPath = d.state.TargetPath
	d.targetDir = d.state.TargetDir
	d.fileLines = d.state.FileLines
	if d.terminalWidth > 0 {
		d.Resize(d.terminalWidth, d.terminalHeight)
	}
}

// --- App model --------------------------------------------------------------

// type model struct {
// 	keys     keymap
// 	content  string // pretend this is your document buffer
// 	filename string // current file path (empty until saved)

// 	// modal state
// 	showingSaveAs bool
// 	dialog        saveDialog

// 	status string
// }
