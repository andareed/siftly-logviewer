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
		view: viewState{
			command: CommandInput{},
		},
	}

	m.view.command.cmd = CmdSort
	m.view.command.buf = "missing-column"
	if got := m.commandPreviewSuffix(); got != " (invalid sort)" {
		t.Fatalf("commandPreviewSuffix sort invalid=%q want=%q", got, " (invalid sort)")
	}

	m.view.command.buf = "1 desc"
	if got := m.commandPreviewSuffix(); got != " (3 matches)" {
		t.Fatalf("commandPreviewSuffix sort valid=%q want=%q", got, " (3 matches)")
	}

	m.view.command.cmd = CmdFilter
	m.view.command.buf = "msg.add"
	if got := m.commandPreviewSuffix(); got != "" {
		t.Fatalf("commandPreviewSuffix filter=%q want empty", got)
	}
}
