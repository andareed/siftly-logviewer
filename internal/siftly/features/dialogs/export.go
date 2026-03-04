package dialogs

import (
	"strings"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Export struct {
	input      textinput.Model
	visible    bool
	lastDir    string
	targetPath string
	targetDir  string
	fileLines  []string
	state      fileDialogState
}

func (d Export) Init() tea.Cmd { return d.input.Focus() }

func NewExportDialog(defaultName, lastDir string) *Export {
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
	d := &Export{input: ti, visible: true, lastDir: lastDir}
	d.refreshPreview()
	return d
}

func (d *Export) Update(msg tea.Msg) (Dialog, Action, tea.Cmd) {
	logging.Debug("ExportDialog:Update:: Called")
	if !d.visible {
		return d, Action{Kind: ActionNone}, nil
	}
	switch m := msg.(type) {
	case tea.KeyMsg:
		logging.Debug("Export:Update::Handle Key Message")
		s := m.String()
		switch s {
		case "enter":
			logging.Infof("ExportDialog:Update::Enter key was pressed, starting exporting to file.")
			d.refreshPreview()
			if !d.state.PrimaryEnabled || d.targetPath == "" {
				return d, Action{Kind: ActionNone}, nil
			}
			return d, Action{Kind: ActionExportConfirm, Path: d.targetPath}, nil
		case "esc":
			logging.Debug("ExportDialog:Update::Esc key was prssed, cancel anything to do with this")
			return d, Action{Kind: ActionExportCancel}, nil
		}
	}
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	d.refreshPreview()
	return d, Action{Kind: ActionNone}, cmd
}

func (d Export) View() string {
	if !d.visible {
		return ""
	}
	innerWidth := 78 - 4
	contentLines := []string{
		dialogSectionLabel("Filename"),
		d.input.View(),
		"",
		dialogSectionLabel("Location"),
		d.targetDir,
		"",
		dialogSectionLabel("Files in folder"),
		strings.Join(d.fileLines, "\n"),
		"",
		dialogStatusLine(d.state.StatusKind, d.state.StatusMessage),
		renderDialogActionRowWithKeys(innerWidth, "Enter", d.state.PrimaryAction, d.state.PrimaryEnabled, "Esc", "Cancel"),
	}
	return renderDialogPanel("Export", dialogTopRightState(d.state.TopRightState), 78, contentLines)
}

func (d *Export) Show() {
	d.visible = true
	d.input.Focus()
}

func (d *Export) Hide() {
	d.visible = false
	d.input.Blur()
}

func (d *Export) Focus() tea.Cmd { return d.input.Focus() }
func (d *Export) Blur()          { d.input.Blur() }
func (d Export) IsVisible() bool { return d.visible }

func (d *Export) refreshPreview() {
	d.state = resolveFileDialogState(d.lastDir, d.input.Value(), d.input.Placeholder, "Export")
	d.targetPath = d.state.TargetPath
	d.targetDir = d.state.TargetDir
	d.fileLines = d.state.FileLines
}
