package siftly

import (
	"strings"
	"testing"
)

func TestCommandHintsLineUsesUniformKeyActionPattern(t *testing.T) {
	m := Model{}
	cmds := []Command{
		CmdJump,
		CmdSearch,
		CmdFilter,
		CmdSort,
		CmdComment,
		CmdMark,
		CmdTimeWindowSet,
		CmdColumns,
		CmdColumnOrder,
	}

	for _, cmd := range cmds {
		hint := m.commandHintsLine(cmd)
		if !strings.Contains(hint, ":") {
			t.Fatalf("hint for %v should contain key/action pairs: %q", cmd, hint)
		}
		if !strings.Contains(hint, "esc: cancel") {
			t.Fatalf("hint for %v should include cancel action: %q", cmd, hint)
		}
	}
}
