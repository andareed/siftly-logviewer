package dialogs

import (
	"fmt"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Help is now just a visible flag + a list of key bindings to show.
type Help struct {
	visible  bool
	bindings []key.Binding
}

func (d Help) Init() tea.Cmd { return nil }

// NewHelpDialog creates a new help dialog showing the given bindings.
func NewHelpDialog(bindings []key.Binding) *Help {
	return &Help{
		visible:  true,
		bindings: bindings,
	}
}

func (d *Help) Update(msg tea.Msg) (Dialog, Action, tea.Cmd) {
	logging.Debug("HelpDialog:Update:: Called")

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "enter", "esc":
			d.visible = false
			return d, Action{Kind: ActionClose}, nil
		}
	}

	return d, Action{Kind: ActionNone}, nil
}

func (d Help) View() string {
	if !d.visible {
		return ""
	}

	// Build lines "keys   description" from the bindings.
	var lines []string
	for _, b := range d.bindings {
		helpItem := b.Help()
		keys, desc := helpItem.Key, helpItem.Desc
		line := fmt.Sprintf("%-12s %s", keys, desc)
		lines = append(lines, line)
	}

	innerWidth := 60 - 4
	contentLines := make([]string, 0, len(lines)+5)
	contentLines = append(contentLines, dialogSectionLabel("Commands"))
	contentLines = append(contentLines, lines...)
	contentLines = append(contentLines, "")
	contentLines = append(contentLines, dialogStatusLine("success", fmt.Sprintf("✓ %d command bindings", len(lines))))
	contentLines = append(contentLines, renderDialogActionRowWithKeys(innerWidth, "Esc", "Close", true, "", ""))
	return renderDialogPanel("Help", dialogTopRightState(fmt.Sprintf("%d commands", len(lines))), 60, contentLines)
}

func (d *Help) Show() {
	d.visible = true
}

func (d *Help) Hide() {
	d.visible = false
}

func (d *Help) Focus() tea.Cmd { return nil }
func (d *Help) Blur()          {}
func (d Help) IsVisible() bool { return d.visible }
