package siftly

import (
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestCommandPreviewSuffix(t *testing.T) {
	m := &Model{
		table: tableState{
			header: []ui.ColumnMeta{
				{Name: "Time", Index: 0, Visible: true},
				{Name: "Message", Index: 1, Visible: true},
			},
			filteredIndices: []int{0, 1, 2},
		},
	}

	m.view.command = newCommandInput(CmdSort, "missing-column")
	if got := m.commandPreviewSuffix(); got != " (invalid sort)" {
		t.Fatalf("commandPreviewSuffix sort invalid=%q want=%q", got, " (invalid sort)")
	}

	m.setCommandValue("1 desc")
	if got := m.commandPreviewSuffix(); got != " (3 matches)" {
		t.Fatalf("commandPreviewSuffix sort valid=%q want=%q", got, " (3 matches)")
	}

	m.view.command = newCommandInput(CmdFilter, "msg.add")
	if got := m.commandPreviewSuffix(); got != "" {
		t.Fatalf("commandPreviewSuffix filter=%q want empty", got)
	}
}
