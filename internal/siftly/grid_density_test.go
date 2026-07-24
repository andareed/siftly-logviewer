package siftly

import (
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

func TestDenseGridRendersModelRowsIntoViewport(t *testing.T) {
	m, err := NewModelFromRecords([][]string{
		{"time", "host", "details"},
		{"10:00", "server-1", "service started"},
		{"10:01", "server-1", "service stopped"},
	}, ColumnSchema{})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.SetStyles(ui.Styles{
		Row:            lipgloss.NewStyle(),
		RowSelected:    lipgloss.NewStyle().Background(lipgloss.Color("#3a3a3a")),
		RepeatedCell:   lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")),
		Cell:           lipgloss.NewStyle().Padding(0, 1),
		RowTextFGColor: lipgloss.Color("#c0c0c0"),
		RowSelectedFG:  lipgloss.Color("#e0e0e0"),
		RowSelectedBG:  lipgloss.Color("#3a3a3a"),
		DefaultMarker:  " ",
		PillMarker:     "▐",
		CommentMarker:  "[*]",
	})
	m.InitialiseView()
	m.applyFilter()
	m.ready = true
	m.terminalWidth = 80
	m.terminalHeight = 24
	m.recomputeLayout(m.terminalHeight, m.terminalWidth)
	renderedRow, _, ok := m.renderRowAt(0)
	if !ok {
		t.Fatal("first row did not render")
	}
	if plainRow := xansi.Strip(renderedRow); !strings.Contains(plainRow, "service started") {
		t.Fatalf("rendered row missing content: %q", plainRow)
	}
	if clipped := xansi.Strip(xansi.Cut(renderedRow, 0, m.viewport.Width)); !strings.Contains(clipped, "service started") {
		t.Fatalf("ANSI-safe viewport clip lost row content (row width %d, viewport %d): %q", xansi.StringWidth(renderedRow), m.viewport.Width, clipped)
	}
	m.refreshView("test", false)

	plain := xansi.Strip(m.viewport.View())
	for _, want := range []string{"1│", "10:00", "server-1", "service started"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("viewport missing %q: %q", want, plain)
		}
	}
}

func TestWrappedGridRecalculatesRowHeightAfterResize(t *testing.T) {
	m, err := NewModelFromRecords([][]string{
		{"details"},
		{"alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"},
	}, ColumnSchema{
		DefaultMinWidth: 8,
		RoleForName:     func(string) ui.ColumnRole { return ui.RolePrimary },
		RoleDefaults: map[ui.ColumnRole]RoleLayout{
			ui.RolePrimary: {MinWidth: 8, Weight: 1, WrapLines: 4},
		},
	})
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	m.SetStyles(ui.Styles{
		Row:            lipgloss.NewStyle(),
		RowSelected:    lipgloss.NewStyle(),
		Cell:           lipgloss.NewStyle().Padding(0, 1),
		RowTextFGColor: lipgloss.Color("#c0c0c0"),
		RowSelectedFG:  lipgloss.Color("#e0e0e0"),
		RowSelectedBG:  lipgloss.Color("#3a3a3a"),
		DefaultMarker:  " ",
	})
	m.InitialiseView()
	m.applyFilter()

	m.recomputeLayout(24, 100)
	_, wideHeight, ok := m.renderRowAt(0)
	if !ok || wideHeight != 1 {
		t.Fatalf("wide row height=%d ok=%t want 1", wideHeight, ok)
	}
	m.recomputeLayout(24, 30)
	_, narrowHeight, ok := m.renderRowAt(0)
	if !ok || narrowHeight <= wideHeight {
		t.Fatalf("narrow row height=%d ok=%t want greater than %d", narrowHeight, ok, wideHeight)
	}
}
