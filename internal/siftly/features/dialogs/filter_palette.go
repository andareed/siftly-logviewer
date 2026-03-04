package dialogs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type filterPaletteTab string

const (
	filterTabPresets filterPaletteTab = "presets"
	filterTabHistory filterPaletteTab = "history"
)

type filterPaletteFocus string

const (
	filterFocusInput filterPaletteFocus = "input"
	filterFocusList  filterPaletteFocus = "list"
)

var lastFilterPaletteTab = filterTabPresets

type FilterPalette struct {
	visible bool
	input   textinput.Model

	presets []FilterPreset
	history []string

	filteredPresets []FilterPreset
	filteredHistory []string

	activeTab filterPaletteTab
	focusArea filterPaletteFocus

	presetCursor  int
	presetScroll  int
	historyCursor int
	historyScroll int

	width      int
	height     int
	selectedFG lipgloss.Color
	selectedBG lipgloss.Color
}

type FilterPreset struct {
	Pattern     string
	Description string
}

func NewFilterPaletteDialog(presets []FilterPreset, history []string, width, height int, selectedFG, selectedBG lipgloss.Color) *FilterPalette {
	w := width - 4
	h := height - 4
	if w < 64 {
		w = 64
	}
	if h < 14 {
		h = 14
	}

	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.Prompt = "Filter: "
	ti.CharLimit = 512
	ti.Width = w - 12

	d := &FilterPalette{
		visible:       true,
		input:         ti,
		presets:       presets,
		history:       history,
		activeTab:     lastFilterPaletteTab,
		focusArea:     filterFocusInput,
		presetCursor:  -1,
		historyCursor: -1,
		width:         w,
		height:        h,
		selectedFG:    selectedFG,
		selectedBG:    selectedBG,
	}
	d.rebuildFiltered()
	d.ensureCursorVisible()
	return d
}

func (d FilterPalette) Init() tea.Cmd { return d.input.Focus() }

func (d *FilterPalette) Update(msg tea.Msg) (Dialog, Action, tea.Cmd) {
	if !d.visible {
		return d, Action{Kind: ActionNone}, nil
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "esc":
			d.visible = false
			return d, Action{Kind: ActionFilterCancel}, nil
		case "ctrl+space":
			if d.focusArea == filterFocusInput {
				d.focusArea = filterFocusList
			} else {
				d.focusArea = filterFocusInput
			}
			return d, Action{Kind: ActionNone}, nil
		case "enter":
			typed := strings.TrimSpace(d.input.Value())
			if typed != "" {
				d.visible = false
				return d, Action{Kind: ActionFilterApply, Pattern: typed}, nil
			}
			if pattern, ok := d.selectedPattern(); ok {
				d.visible = false
				return d, Action{Kind: ActionFilterApply, Pattern: pattern}, nil
			}
			return d, Action{Kind: ActionNone}, nil
		case "up", "ctrl+p":
			d.focusArea = filterFocusList
			d.move(-1)
			return d, Action{Kind: ActionNone}, nil
		case "down", "ctrl+n":
			d.focusArea = filterFocusList
			d.move(1)
			return d, Action{Kind: ActionNone}, nil
		case "tab":
			d.switchTab(1)
			return d, Action{Kind: ActionNone}, nil
		case "shift+tab":
			d.switchTab(-1)
			return d, Action{Kind: ActionNone}, nil
		case "pgup":
			d.focusArea = filterFocusList
			d.movePage(-1)
			return d, Action{Kind: ActionNone}, nil
		case "pgdown":
			d.focusArea = filterFocusList
			d.movePage(1)
			return d, Action{Kind: ActionNone}, nil
		case "h":
			if d.focusArea == filterFocusList {
				d.switchTab(-1)
				return d, Action{Kind: ActionNone}, nil
			}
		case "j":
			if d.focusArea == filterFocusList {
				d.move(1)
				return d, Action{Kind: ActionNone}, nil
			}
		case "k":
			if d.focusArea == filterFocusList {
				d.move(-1)
				return d, Action{Kind: ActionNone}, nil
			}
		case "l":
			if d.focusArea == filterFocusList {
				d.switchTab(1)
				return d, Action{Kind: ActionNone}, nil
			}
		}

		if d.focusArea == filterFocusList {
			return d, Action{Kind: ActionNone}, nil
		}
	}

	prev := d.input.Value()
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	if d.input.Value() != prev {
		d.rebuildFiltered()
		d.ensureCursorVisible()
	}
	return d, Action{Kind: ActionNone}, cmd
}

func (d FilterPalette) View() string {
	if !d.visible {
		return ""
	}
	innerWidth := d.width - 4
	if innerWidth < 12 {
		innerWidth = 12
	}

	matchCount := len(d.filteredPresets)
	if d.activeTab == filterTabHistory {
		matchCount = len(d.filteredHistory)
	}
	topRight := dialogTopRightState(fmt.Sprintf("%d matches", matchCount))

	tabs := d.renderTabs(innerWidth)
	focusLabel := "Input"
	if d.focusArea == filterFocusList {
		focusLabel = "List"
	}
	activeTabLabel := "Presets"
	if d.activeTab == filterTabHistory {
		activeTabLabel = "History"
	}
	selected, hasSelected := d.selectedPattern()
	typed := strings.TrimSpace(d.input.Value())
	primaryEnabled := strings.TrimSpace(typed) != "" || hasSelected
	statusMsg := fmt.Sprintf("✓ %d matches in %s", matchCount, activeTabLabel)
	if matchCount == 0 {
		statusMsg = "✖ No matches"
	}
	if typed == "" && hasSelected {
		statusMsg = "✓ Selected: " + truncate(selected, max(16, innerWidth-12))
	}

	contentLines := []string{
		dialogSectionLabel("Query"),
		d.input.View(),
		"",
		dialogStatusLine(func() string {
			if matchCount == 0 {
				return "error"
			}
			return "success"
		}(), statusMsg),
		renderDialogActionRowWithKeys(innerWidth, "Enter", "Apply", primaryEnabled, "Esc", "Cancel"),
		"",
		dialogSectionLabel("List"),
		tabs,
		lipgloss.NewStyle().Faint(true).Render("Focus: " + focusLabel + "  Ctrl+Space: toggle  Tab: switch tabs  ↑/↓/Pg: move"),
		strings.Join(d.renderActiveList(innerWidth, d.listPanelContentHeight()), "\n"),
	}

	return renderDialogPanel("Filter Palette", topRight, d.width, contentLines)
}

func (d *FilterPalette) Show() {
	d.visible = true
	d.input.Focus()
}

func (d *FilterPalette) Hide() {
	d.visible = false
	d.input.Blur()
}

func (d *FilterPalette) Focus() tea.Cmd { return d.input.Focus() }
func (d *FilterPalette) Blur()          { d.input.Blur() }
func (d FilterPalette) IsVisible() bool { return d.visible }
