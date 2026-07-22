package dialogs

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
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
	listHeight int
	selectedFG lipgloss.Color
	selectedBG lipgloss.Color
	tokens     ui.DesignTokens
}

type FilterPreset struct {
	Pattern     string
	Description string
}

func NewFilterPaletteDialog(presets []FilterPreset, history []string, width, height int, selectedFG, selectedBG lipgloss.Color, tokenOptions ...ui.DesignTokens) *FilterPalette {
	return newFilterPaletteDialog(presets, history, width, height, selectedFG, selectedBG, lastFilterPaletteTab, tokenOptions...)
}

func NewFilterHistoryPaletteDialog(presets []FilterPreset, history []string, width, height int, selectedFG, selectedBG lipgloss.Color, tokenOptions ...ui.DesignTokens) *FilterPalette {
	return newFilterPaletteDialog(presets, history, width, height, selectedFG, selectedBG, filterTabHistory, tokenOptions...)
}

func newFilterPaletteDialog(presets []FilterPreset, history []string, width, height int, selectedFG, selectedBG lipgloss.Color, initialTab filterPaletteTab, tokenOptions ...ui.DesignTokens) *FilterPalette {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.Prompt = "Filter: "
	ti.CharLimit = 512
	tokens := dialogDesignTokens(tokenOptions)
	styleDialogTextInput(&ti, tokens)

	d := &FilterPalette{
		visible:       true,
		input:         ti,
		presets:       presets,
		history:       history,
		activeTab:     initialTab,
		focusArea:     filterFocusInput,
		presetCursor:  -1,
		historyCursor: -1,
		selectedFG:    selectedFG,
		selectedBG:    selectedBG,
		tokens:        tokens,
	}
	d.rebuildFiltered()
	d.Resize(width, height)
	d.ensureCursorVisible()
	return d
}

func (d *FilterPalette) Resize(terminalWidth, terminalHeight int) {
	d.width = responsiveDialogWidth(terminalWidth, d.preferredWidth(), 48)
	d.height = responsiveDialogHeight(terminalHeight, 28)
	d.listHeight = boundedListRows(d.height, 11, -1, 17)
	d.input.Width = max(1, d.width-12)
	d.ensureCursorVisible()
}

func (d FilterPalette) preferredWidth() int {
	innerWidth := lipgloss.Width("Focus: Input  Ctrl+Space: toggle  Tab: switch tabs")
	for _, preset := range d.presets {
		candidate := max(lipgloss.Width(preset.Description), lipgloss.Width(preset.Pattern)+4)
		if candidate > innerWidth {
			innerWidth = candidate
		}
	}
	for _, history := range d.history {
		if candidate := lipgloss.Width(history) + 2; candidate > innerWidth {
			innerWidth = candidate
		}
	}
	return clampDialogWidth(innerWidth+4, 48, 96)
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
	topRight := dialogTopRightState(fmt.Sprintf("%d matches", matchCount), d.tokens)

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
		dialogSectionLabel("Query", d.tokens),
		d.input.View(),
		"",
		dialogStatusLine(func() string {
			if matchCount == 0 {
				return "error"
			}
			return "success"
		}(), statusMsg, d.tokens),
		renderDialogActionRowWithKeys(innerWidth, "Enter", "Apply", primaryEnabled, "Esc", "Cancel", d.tokens),
		"",
		dialogSectionLabel("List", d.tokens),
		tabs,
		d.tokens.Emphasis.Muted.Render("Focus: " + focusLabel + "  Ctrl+Space: toggle  Tab: switch tabs  ↑/↓/Pg: move"),
		strings.Join(d.renderActiveList(innerWidth, d.listPanelContentHeight()), "\n"),
	}

	return renderDialogPanel("Filter Palette", topRight, d.width, contentLines, d.tokens)
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
