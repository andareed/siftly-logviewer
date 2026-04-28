package siftly

import (
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
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
	if m.view.command.buf != "existing note" {
		t.Fatalf("comment seed mismatch: got %q want %q", m.view.command.buf, "existing note")
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
	if m.view.mainBodySnapshotWidth != m.panelWidth() {
		t.Fatalf("snapshot width = %d want %d", m.view.mainBodySnapshotWidth, m.panelWidth())
	}
	if m.view.mainBodySnapshotHeight != m.terminalHeight {
		t.Fatalf("snapshot height = %d want %d", m.view.mainBodySnapshotHeight, m.terminalHeight)
	}

	snapshot := m.view.mainBodySnapshot
	m.view.command.buf = "abc"
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
}
