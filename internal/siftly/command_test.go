package siftly

import "testing"

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
