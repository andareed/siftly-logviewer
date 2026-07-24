package dialogs

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Recovery asks the user what to do with an unsaved sidecar found at startup.
type Recovery struct {
	visible        bool
	sourceName     string
	savedAt        string
	contents       string
	width          int
	terminalWidth  int
	terminalHeight int
	tokens         ui.DesignTokens
}

func NewRecoveryDialog(sourceName, savedAt, contents string, terminalWidth, terminalHeight int, tokenOptions ...ui.DesignTokens) *Recovery {
	d := &Recovery{
		visible:    true,
		sourceName: strings.TrimSpace(sourceName),
		savedAt:    strings.TrimSpace(savedAt),
		contents:   strings.TrimSpace(contents),
		tokens:     dialogDesignTokens(tokenOptions),
	}
	d.Resize(terminalWidth, terminalHeight)
	return d
}

func (d Recovery) Init() tea.Cmd { return nil }

func (d *Recovery) Update(msg tea.Msg) (Dialog, Action, tea.Cmd) {
	if !d.visible {
		return d, Action{Kind: ActionNone}, nil
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch strings.ToLower(keyMsg.String()) {
		case "r":
			return d, Action{Kind: ActionRecoveryRestore}, nil
		case "d":
			return d, Action{Kind: ActionRecoveryDiscard}, nil
		case "q", "ctrl+c":
			return d, Action{Kind: ActionRecoveryQuit}, nil
		}
	}
	return d, Action{Kind: ActionNone}, nil
}

func (d Recovery) View() string {
	if !d.visible {
		return ""
	}

	innerWidth := max(1, d.width-4)
	content := []string{
		"Unsaved changes from an earlier session were found.",
		"",
		recoveryField("File", d.sourceName, d.tokens),
		recoveryField("Saved", d.savedAt, d.tokens),
		recoveryField("Contains", d.contents, d.tokens),
		"",
		dialogStatusLine("warn", "Choose whether to restore or discard them.", d.tokens),
		renderDialogActionRowWithKeys(innerWidth, "r", "Restore", true, "d", "Discard", d.tokens),
		d.tokens.Emphasis.Muted.Render("q Quit and leave recovery untouched"),
	}
	return renderDialogPanel("Recovery Found", "DECISION REQUIRED", d.width, content, d.tokens)
}

func (d *Recovery) Show()          { d.visible = true }
func (d *Recovery) Hide()          { d.visible = false }
func (d *Recovery) Focus() tea.Cmd { return nil }
func (d *Recovery) Blur()          {}
func (d Recovery) IsVisible() bool { return d.visible }

func (d *Recovery) Resize(terminalWidth, terminalHeight int) {
	d.terminalWidth = terminalWidth
	d.terminalHeight = terminalHeight

	preferred := lipgloss.Width("Unsaved changes from an earlier session were found.") + 4
	for _, value := range []string{d.sourceName, d.savedAt, d.contents} {
		if width := lipgloss.Width(value) + 14; width > preferred {
			preferred = width
		}
	}
	d.width = responsiveDialogWidth(terminalWidth, clampDialogWidth(preferred, 48, 76), 36)
}

func recoveryField(label, value string, tokens ui.DesignTokens) string {
	if value == "" {
		value = "unknown"
	}
	return fmt.Sprintf("%s  %s", dialogSectionLabel(label, tokens), value)
}
