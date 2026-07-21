package siftly

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOutputKeysUseSaveAndExportPrefix(t *testing.T) {
	m := newChangeTrackingTestModel()

	tests := []struct {
		name string
		key  rune
		want viewAction
	}{
		{name: "save Siftly JSON", key: 's', want: viewActionSave},
		{name: "export operations", key: 'e', want: viewActionPrefixExport},
		{name: "old graph export", key: 'W', want: viewActionNone},
		{name: "old reload", key: 'R', want: viewActionNone},
	}
	for _, tt := range tests {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
		if got := m.resolveViewAction(msg); got != tt.want {
			t.Fatalf("%s action = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestExportPrefixRoutesDataAndConditionalGraph(t *testing.T) {
	m := newChangeTrackingTestModel()

	m.view.pendingViewPrefix = "e"
	handled, cmd, _ := m.handleViewPrefixKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !handled || cmd != nil || m.activeDialog == nil || !m.activeDialog.IsVisible() {
		t.Fatalf("e d should open the filtered data export dialog: handled=%t cmd=%v dialog=%T", handled, cmd != nil, m.activeDialog)
	}

	m.activeDialog = nil
	m.view.pendingViewPrefix = "e"
	handled, cmd, _ = m.handleViewPrefixKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if !handled || cmd == nil || !strings.Contains(m.view.notice.Msg, "not available") {
		t.Fatalf("e g without graph support should report unavailable: handled=%t cmd=%v notice=%q", handled, cmd != nil, m.view.notice.Msg)
	}

	m.graphConfig.Enabled = true
	m.view.pendingViewPrefix = "e"
	handled, cmd, _ = m.handleViewPrefixKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if !handled || cmd == nil {
		t.Fatalf("e g with graph support should start graph export: handled=%t cmd=%v", handled, cmd != nil)
	}
}

func TestExportHintsOnlyShowGraphWhenEnabled(t *testing.T) {
	m := newChangeTrackingTestModel()
	m.view.pendingViewPrefix = "e"
	if got := m.footerHints(false, CmdNone); strings.Contains(got, "graph") {
		t.Fatalf("graph-disabled export hint includes graph: %q", got)
	}

	m.graphConfig.Enabled = true
	if got := m.footerHints(false, CmdNone); !strings.Contains(got, "g: graph SVG") {
		t.Fatalf("graph-enabled export hint omits graph: %q", got)
	}
}

func TestExportHelpOnlyShowsGraphWhenEnabled(t *testing.T) {
	for _, tt := range []struct {
		name         string
		graphEnabled bool
		wantGraph    bool
	}{
		{name: "graph disabled", graphEnabled: false, wantGraph: false},
		{name: "graph enabled", graphEnabled: true, wantGraph: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := newChangeTrackingTestModel()
			m.graphConfig.Enabled = tt.graphEnabled
			foundData := false
			foundGraph := false
			for _, item := range m.commandItems() {
				switch item.ID {
				case commandExportData:
					foundData = item.Shortcut == "e d"
				case commandExportGraph:
					foundGraph = item.Shortcut == "e g"
				}
			}
			if !foundData {
				t.Fatal("command registry omits filtered data export")
			}
			if foundGraph != tt.wantGraph {
				t.Fatalf("graph export present = %t, want %t", foundGraph, tt.wantGraph)
			}
		})
	}
}
