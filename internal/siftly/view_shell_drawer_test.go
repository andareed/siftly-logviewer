package siftly

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

func TestCommentDrawerViewMatchesPanelWidth(t *testing.T) {
	const panelWidth = 96

	m := Model{
		cursor: 0,
		table: tableState{
			rows: []Row{
				{ID: 42},
			},
			filteredIndices: []int{0},
		},
	}
	m.drawerPort = viewport.New(0, 0)
	m.drawerPort.Width = panelWidth - 4
	m.drawerPort.Height = drawerContentRows
	m.table.commentRows = map[uint64]string{42: "example note"}
	m.drawerPort.SetContent("example note")

	got := stripANSI(m.commentDrawerView(panelWidth))
	lines := strings.Split(got, "\n")

	wantLines := drawerContentRows + drawerChromeRows
	if len(lines) != wantLines {
		t.Fatalf("drawer line count mismatch: got %d want %d", len(lines), wantLines)
	}
	for i, line := range lines {
		if lipgloss.Width(line) != panelWidth {
			t.Fatalf("line %d width mismatch: got %d want %d (%q)", i, lipgloss.Width(line), panelWidth, line)
		}
	}
	if !strings.Contains(lines[0], "Comment") {
		t.Fatalf("drawer top border missing title: %q", lines[0])
	}
	if !strings.Contains(lines[0], "Chars 12") {
		t.Fatalf("drawer top border missing char status: %q", lines[0])
	}
}
