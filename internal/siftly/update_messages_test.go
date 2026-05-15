package siftly

import (
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

type recordingDialog struct {
	visible bool
	updates []tea.Msg
}

func (d *recordingDialog) Init() tea.Cmd { return nil }

func (d *recordingDialog) Update(msg tea.Msg) (dialogs.Dialog, dialogs.Action, tea.Cmd) {
	d.updates = append(d.updates, msg)
	return d, dialogs.Action{Kind: dialogs.ActionNone}, nil
}

func (d *recordingDialog) View() string    { return "" }
func (d *recordingDialog) Focus() tea.Cmd  { return nil }
func (d *recordingDialog) Blur()           {}
func (d *recordingDialog) IsVisible() bool { return d.visible }
func (d *recordingDialog) Show()           { d.visible = true }
func (d *recordingDialog) Hide()           { d.visible = false }

func TestHandleDialogInputForwardsNonKeyMessages(t *testing.T) {
	dialog := &recordingDialog{visible: true}
	m := &Model{activeDialog: dialog}
	msg := struct{ text string }{text: "paste"}

	cmd, handled := m.handleDialogInput(msg)
	if !handled {
		t.Fatalf("dialog input should handle non-key messages")
	}
	if cmd != nil {
		t.Fatalf("dialog input should not return a command")
	}
	if len(dialog.updates) != 1 || dialog.updates[0] != msg {
		t.Fatalf("dialog updates = %#v, want forwarded message %#v", dialog.updates, msg)
	}
}

func TestHandleDialogInputLeavesResizeMessagesForWindowHandler(t *testing.T) {
	dialog := &recordingDialog{visible: true}
	m := &Model{activeDialog: dialog}

	cmd, handled := m.handleDialogInput(tea.WindowSizeMsg{Width: 80, Height: 24})
	if handled {
		t.Fatalf("resize messages should not be consumed by dialog input")
	}
	if cmd != nil {
		t.Fatalf("resize pass-through should not return a command")
	}
	if len(dialog.updates) != 0 {
		t.Fatalf("resize message should not be forwarded to dialog")
	}
}
