package siftly

import (
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestCommentDrawerTracksWrappedContentHeight(t *testing.T) {
	m := inspectorTestModel()
	m.SetStyles(ui.DefaultStyles())
	m.view.drawerOpen = true
	m.recomputeLayout(40, 100)

	if got := m.drawerPort.Height; got != 1 {
		t.Fatalf("empty drawer content height = %d, want 1", got)
	}
	emptyViewportHeight := m.viewport.Height

	m.table.commentRows[m.table.rows[0].ID] = strings.Repeat("diagnostic note ", 35)
	m.recomputeLayout(40, 100)
	if got := m.drawerPort.Height; got <= 1 || got > drawerContentRows {
		t.Fatalf("wrapped drawer content height = %d, want 2..%d", got, drawerContentRows)
	}
	if m.viewport.Height >= emptyViewportHeight {
		t.Fatalf("table viewport did not yield space to longer drawer: before=%d after=%d", emptyViewportHeight, m.viewport.Height)
	}
}

func TestGraphHeightRespondsToTerminalHeight(t *testing.T) {
	m := inspectorTestModel()
	m.SetStyles(ui.DefaultStyles())
	m.graphConfig.Enabled = true
	m.view.graphWindow.Open = true
	m.view.graphWindow.Height = 16

	m.recomputeLayout(24, 100)
	if got, want := m.view.graphHeight, 8; got != want {
		t.Fatalf("compact graph height = %d, want %d", got, want)
	}
	if m.viewport.Height < minimumTableRows {
		t.Fatalf("compact graph left only %d table rows", m.viewport.Height)
	}

	m.recomputeLayout(40, 100)
	if got, want := m.view.graphHeight, 16; got != want {
		t.Fatalf("wide graph height = %d, want configured %d", got, want)
	}
}

func TestGraphAndInspectorShareBodyWithoutStarvingTable(t *testing.T) {
	m := inspectorTestModel()
	m.SetStyles(ui.DefaultStyles())
	m.graphConfig.Enabled = true
	m.view.graphWindow.Open = true
	m.view.graphWindow.Height = 16
	m.view.inspector.open = true
	m.recomputeLayout(24, 100)

	bodyHeight := 24 - m.styles.App.GetVerticalFrameSize() - panelChromeRows - footerRows
	graphOuterHeight := m.view.graphHeight + m.styles.GraphArea.GetVerticalFrameSize()
	used := m.viewport.Height + graphOuterHeight + m.view.inspector.height
	if used != bodyHeight {
		t.Fatalf("responsive body allocation = %d, want %d", used, bodyHeight)
	}
	if m.viewport.Height < minimumTableRows {
		t.Fatalf("stacked panels left only %d table rows", m.viewport.Height)
	}
	if m.inspectorPort.Height < 1 || m.view.graphHeight < 1 {
		t.Fatalf("stacked panels were not both visible: inspector=%d graph=%d", m.inspectorPort.Height, m.view.graphHeight)
	}
}

func TestTimeWindowOverlaySizingStaysInsideCompactTerminal(t *testing.T) {
	if got, want := responsiveOverlayWidth(48, 104, 52), 44; got != want {
		t.Fatalf("compact overlay width = %d, want %d", got, want)
	}
	lines := []string{"window", "scrubber", "", "range", "start", "end", "", "status", "actions", "controls"}
	if got := len(fitTimeWindowDialogLines(lines, 6)); got != 6 {
		t.Fatalf("compact time-window content lines = %d, want 6", got)
	}
}
