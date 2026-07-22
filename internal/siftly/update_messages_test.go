package siftly

import (
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	tea "github.com/charmbracelet/bubbletea"
)

type recordingDialog struct {
	visible      bool
	updates      []tea.Msg
	resizeWidth  int
	resizeHeight int
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
func (d *recordingDialog) Resize(width, height int) {
	d.resizeWidth = width
	d.resizeHeight = height
}

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

func TestWindowHandlerResizesOpenDialog(t *testing.T) {
	dialog := &recordingDialog{visible: true}
	m := &Model{activeDialog: dialog}

	_, handled := m.handleWindowMsg(tea.WindowSizeMsg{Width: 72, Height: 20})
	if !handled {
		t.Fatal("window resize was not handled")
	}
	if dialog.resizeWidth != 72 || dialog.resizeHeight != 20 {
		t.Fatalf("dialog resized to %dx%d, want 72x20", dialog.resizeWidth, dialog.resizeHeight)
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
