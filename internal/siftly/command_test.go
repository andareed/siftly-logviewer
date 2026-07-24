package siftly

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestEnterCommandCommentSeedsExistingComment(t *testing.T) {
	rowID := uint64(42)
	m := &Model{
		cursor: 0,
		table: tableState{
			rows:            []Row{{ID: rowID, Cols: []string{"a"}}},
			filteredIndices: []int{0},
			commentRows: map[uint64]string{
				rowID: "existing note",
			},
		},
	}

	_ = m.enterCommand(CmdComment, "", false, false)

	if m.view.mode != modeCommand {
		t.Fatalf("mode not set to command: got %v", m.view.mode)
	}
	if m.view.command.cmd != CmdComment {
		t.Fatalf("command not set to comment: got %v", m.view.command.cmd)
	}
	if got := m.commandValue(); got != "existing note" {
		t.Fatalf("comment seed mismatch: got %q want %q", got, "existing note")
	}
}

func TestEnterFooterTextCommandCapturesMainBodySnapshot(t *testing.T) {
	m := &Model{
		ready:          true,
		terminalHeight: 24,
		viewport:       viewport.New(20, 5),
		cursor:         0,
		table: tableState{
			rows:            []Row{{ID: 42, Cols: []string{"alpha"}, OriginalIndex: 1}},
			filteredIndices: []int{0},
		},
	}
	m.viewport.SetContent("alpha")

	_ = m.enterCommand(CmdFilter, "", false, false)

	if !m.view.mainBodySnapshotActive {
		t.Fatalf("main body snapshot should be active for footer text command")
	}
	if m.view.mainBodySnapshot == "" {
		t.Fatalf("main body snapshot should be captured")
	}
	if m.view.mainBodyFrameSnapshot == "" {
		t.Fatalf("main body frame snapshot should be captured")
	}
	if m.view.mainBodySnapshotWidth != m.panelWidth() {
		t.Fatalf("snapshot width = %d want %d", m.view.mainBodySnapshotWidth, m.panelWidth())
	}
	if m.view.mainBodySnapshotHeight != m.terminalHeight {
		t.Fatalf("snapshot height = %d want %d", m.view.mainBodySnapshotHeight, m.terminalHeight)
	}

	snapshot := m.view.mainBodySnapshot
	m.setCommandValue("abc")
	if got := m.mainBodyForView(m.panelWidth()); got != snapshot {
		t.Fatalf("main body should reuse snapshot while footer text changes")
	}
}

func TestExitCommandClearsMainBodySnapshot(t *testing.T) {
	m := &Model{
		ready:          true,
		terminalHeight: 24,
		viewport:       viewport.New(20, 5),
	}

	_ = m.enterCommand(CmdSearch, "", false, false)
	if !m.view.mainBodySnapshotActive {
		t.Fatalf("main body snapshot should be active")
	}

	_ = m.exitCommand(false)

	if m.view.mainBodySnapshotActive {
		t.Fatalf("main body snapshot should be inactive after command exit")
	}
	if m.view.mainBodySnapshot != "" {
		t.Fatalf("main body snapshot should be cleared")
	}
	if m.view.mainBodyFrameSnapshot != "" {
		t.Fatalf("main body frame snapshot should be cleared")
	}
}

func TestSnapshotAppFrameMatchesLiveAppFrame(t *testing.T) {
	m := &Model{
		ready:          true,
		terminalHeight: 24,
		viewport:       viewport.New(20, 5),
		styles:         ui.Styles{App: lipgloss.NewStyle().Margin(1, 2)},
		cursor:         0,
		table: tableState{
			rows:            []Row{{ID: 42, Cols: []string{"alpha"}, OriginalIndex: 1}},
			filteredIndices: []int{0},
		},
	}
	m.viewport.SetContent("alpha")

	_ = m.enterCommand(CmdFilter, "", false, false)
	snapshotFrame := m.View()

	m.view.mainBodySnapshotActive = false
	m.view.mainBodySnapshot = ""
	m.view.mainBodyFrameSnapshot = ""
	liveFrame := m.View()

	if snapshotFrame != liveFrame {
		t.Fatalf("snapshot frame should match live frame")
	}
}

func TestCommandInputAcceptsBatchedRunes(t *testing.T) {
	m := &Model{}
	_ = m.enterCommand(CmdFilter, "", false, false)

	_, cmd := m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("va.msg")})
	if cmd != nil {
		t.Fatalf("batched runes should not return command")
	}
	if got := m.commandValue(); got != "va.msg" {
		t.Fatalf("command value = %q, want %q", got, "va.msg")
	}
}

func TestCommandInputUsesTextInputCursorEditing(t *testing.T) {
	m := &Model{}
	_ = m.enterCommand(CmdFilter, "", false, false)

	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ab")})
	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyLeft})
	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("X")})
	_, _ = m.handleCommandKey(tea.KeyMsg{Type: tea.KeyBackspace})

	if got := m.commandValue(); got != "ab" {
		t.Fatalf("command value after cursor editing = %q, want %q", got, "ab")
	}
}

func TestCommandInputCtrlVReturnsPasteCommand(t *testing.T) {
	m := &Model{}
	_ = m.enterCommand(CmdFilter, "", false, false)

	_, cmd := m.handleCommandKey(tea.KeyMsg{Type: tea.KeyCtrlV})
	if cmd == nil {
		t.Fatalf("ctrl+v should return textinput paste command")
	}
}

func TestFilterCommandCtrlRAndUpOpenHistoryPalette(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, "history.json")
	if err := os.WriteFile(historyPath, []byte(`{"history":["history-pattern","older-pattern"]}`), 0o644); err != nil {
		t.Fatalf("write history: %v", err)
	}

	m := &Model{
		terminalWidth:  100,
		terminalHeight: 30,
	}
	_ = m.enterCommand(CmdFilter, "", false, false)
	m.SetFilterConfig(FilterConfigSettings{
		DefaultPresets: PresetList{{Pattern: "preset-pattern", Description: "preset"}},
		HistoryPath:    historyPath,
	})

	for _, keyMsg := range []tea.KeyMsg{
		{Type: tea.KeyCtrlR},
		{Type: tea.KeyUp},
	} {
		m.activeDialog = nil
		_, cmd := m.handleCommandKey(keyMsg)
		if cmd != nil {
			t.Fatalf("%s should open history palette without returning a command", keyMsg.String())
		}
		if m.activeDialog == nil || !m.activeDialog.IsVisible() {
			t.Fatalf("%s should open the filter palette", keyMsg.String())
		}

		_, action, _ := m.activeDialog.Update(tea.KeyMsg{Type: tea.KeyEnter})
		if action.Kind != dialogs.ActionFilterApply {
			t.Fatalf("%s enter action = %v, want ActionFilterApply", keyMsg.String(), action.Kind)
		}
		if action.Pattern != "history-pattern" {
			t.Fatalf("%s selected pattern = %q, want %q", keyMsg.String(), action.Pattern, "history-pattern")
		}
	}
}
