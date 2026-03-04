package dialogs

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/charmbracelet/lipgloss"
)

func (d *FilterPalette) rebuildFiltered() {
	query := strings.ToLower(strings.TrimSpace(d.input.Value()))
	d.filteredPresets = filterPresets(d.presets, query)
	d.filteredHistory = filterValues(d.history, query)
	d.clampTabCursors()
}

func (d *FilterPalette) clampTabCursors() {
	if len(d.filteredPresets) == 0 {
		d.presetCursor = -1
		d.presetScroll = 0
	} else {
		if d.presetCursor < 0 {
			d.presetCursor = 0
		}
		if d.presetCursor >= len(d.filteredPresets) {
			d.presetCursor = len(d.filteredPresets) - 1
		}
		if d.presetScroll < 0 {
			d.presetScroll = 0
		}
	}

	if len(d.filteredHistory) == 0 {
		d.historyCursor = -1
		d.historyScroll = 0
	} else {
		if d.historyCursor < 0 {
			d.historyCursor = 0
		}
		if d.historyCursor >= len(d.filteredHistory) {
			d.historyCursor = len(d.filteredHistory) - 1
		}
		if d.historyScroll < 0 {
			d.historyScroll = 0
		}
	}

	if d.activeTab != filterTabPresets && d.activeTab != filterTabHistory {
		d.activeTab = filterTabPresets
	}
}

func (d *FilterPalette) switchTab(delta int) {
	if delta == 0 {
		return
	}
	if delta > 0 {
		if d.activeTab == filterTabPresets {
			d.activeTab = filterTabHistory
		} else {
			d.activeTab = filterTabPresets
		}
	} else {
		if d.activeTab == filterTabHistory {
			d.activeTab = filterTabPresets
		} else {
			d.activeTab = filterTabHistory
		}
	}
	lastFilterPaletteTab = d.activeTab
	d.ensureCursorVisible()
}

func (d *FilterPalette) selectedPattern() (string, bool) {
	switch d.activeTab {
	case filterTabPresets:
		if d.presetCursor < 0 || d.presetCursor >= len(d.filteredPresets) {
			return "", false
		}
		return d.filteredPresets[d.presetCursor].Pattern, true
	case filterTabHistory:
		if d.historyCursor < 0 || d.historyCursor >= len(d.filteredHistory) {
			return "", false
		}
		return d.filteredHistory[d.historyCursor], true
	default:
		return "", false
	}
}

func (d *FilterPalette) move(delta int) {
	if delta == 0 {
		return
	}
	switch d.activeTab {
	case filterTabPresets:
		if len(d.filteredPresets) == 0 {
			return
		}
		next := d.presetCursor + delta
		if next < 0 {
			next = 0
		}
		if next >= len(d.filteredPresets) {
			next = len(d.filteredPresets) - 1
		}
		d.presetCursor = next
	case filterTabHistory:
		if len(d.filteredHistory) == 0 {
			return
		}
		next := d.historyCursor + delta
		if next < 0 {
			next = 0
		}
		if next >= len(d.filteredHistory) {
			next = len(d.filteredHistory) - 1
		}
		d.historyCursor = next
	}
	d.ensureCursorVisible()
}

func (d *FilterPalette) movePage(delta int) {
	step := d.visibleSlots()
	if step < 1 {
		step = 1
	}
	d.move(delta * step)
}

func (d *FilterPalette) listPanelContentHeight() int {
	innerHeight := d.height - 4
	if innerHeight < 5 {
		innerHeight = 5
	}
	panelHeight := innerHeight - 8
	if panelHeight < 3 {
		panelHeight = 3
	}
	return panelHeight
}

func (d *FilterPalette) visibleSlots() int {
	h := d.listPanelContentHeight()
	if d.activeTab == filterTabPresets {
		slots := (h + 1) / 4
		if slots < 1 {
			return 1
		}
		return slots
	}
	slots := (h + 1) / 2
	if slots < 1 {
		return 1
	}
	return slots
}

func (d *FilterPalette) ensureCursorVisible() {
	slots := d.visibleSlots()
	if slots < 1 {
		slots = 1
	}

	switch d.activeTab {
	case filterTabPresets:
		if d.presetCursor < 0 {
			d.presetScroll = 0
			return
		}
		if d.presetCursor < d.presetScroll {
			d.presetScroll = d.presetCursor
			return
		}
		if d.presetCursor >= d.presetScroll+slots {
			d.presetScroll = d.presetCursor - slots + 1
			return
		}
	case filterTabHistory:
		if d.historyCursor < 0 {
			d.historyScroll = 0
			return
		}
		if d.historyCursor < d.historyScroll {
			d.historyScroll = d.historyCursor
			return
		}
		if d.historyCursor >= d.historyScroll+slots {
			d.historyScroll = d.historyCursor - slots + 1
			return
		}
	}
}

func (d FilterPalette) renderTabs(width int) string {
	presets := fmt.Sprintf(" Presets (%d) ", len(d.filteredPresets))
	history := fmt.Sprintf(" History (%d) ", len(d.filteredHistory))

	active := lipgloss.NewStyle().Bold(true).Underline(true)
	inactive := lipgloss.NewStyle().Faint(true)

	if d.activeTab == filterTabPresets {
		presets = active.Render(presets)
		history = inactive.Render(history)
	} else {
		presets = inactive.Render(presets)
		history = active.Render(history)
	}

	line := presets + " " + history
	return truncate(line, width)
}

func (d FilterPalette) renderActiveList(width, maxLines int) []string {
	lines := make([]string, 0, maxLines)

	switch d.activeTab {
	case filterTabPresets:
		if len(d.filteredPresets) == 0 {
			return []string{lipgloss.NewStyle().Faint(true).Render("No preset matches")}
		}
		slots := d.visibleSlots()
		end := d.presetScroll + slots
		if end > len(d.filteredPresets) {
			end = len(d.filteredPresets)
		}
		for i := d.presetScroll; i < end; i++ {
			item := d.filteredPresets[i]
			desc := strings.TrimSpace(item.Description)
			if desc == "" {
				desc = "(no description)"
			}

			prefix := "  "
			if i == d.presetCursor {
				prefix = "> "
			}

			descLine := prefix + truncate(desc, width-2)
			patternLine1, patternLine2 := wrapPatternTwoLines(item.Pattern, width-4)
			patternLine1 = "    " + patternLine1
			patternLine2 = "    " + patternLine2

			if i == d.presetCursor {
				style := lipgloss.NewStyle().Bold(true)
				if d.selectedFG != "" {
					style = style.Foreground(d.selectedFG)
				}
				if d.selectedBG != "" {
					style = style.Background(d.selectedBG)
				}
				lines = append(lines, style.Render(descLine))
				lines = append(lines, style.Render(patternLine1))
				lines = append(lines, style.Render(patternLine2))
			} else {
				lines = append(lines, descLine)
				lines = append(lines, lipgloss.NewStyle().Faint(true).Render(patternLine1))
				lines = append(lines, lipgloss.NewStyle().Faint(true).Render(patternLine2))
			}
			if i < end-1 {
				lines = append(lines, lipgloss.NewStyle().Faint(true).Render(ruleLine(width)))
			}
		}
	case filterTabHistory:
		if len(d.filteredHistory) == 0 {
			return []string{lipgloss.NewStyle().Faint(true).Render("No history matches")}
		}
		slots := d.visibleSlots()
		end := d.historyScroll + slots
		if end > len(d.filteredHistory) {
			end = len(d.filteredHistory)
		}
		for i := d.historyScroll; i < end; i++ {
			line := "  " + truncate(d.filteredHistory[i], width-2)
			rendered := line
			if i == d.historyCursor {
				line = "> " + truncate(d.filteredHistory[i], width-2)
				style := lipgloss.NewStyle().Bold(true)
				if d.selectedFG != "" {
					style = style.Foreground(d.selectedFG)
				}
				if d.selectedBG != "" {
					style = style.Background(d.selectedBG)
				}
				rendered = style.Render(line)
			}
			lines = append(lines, rendered)
			if i < end-1 {
				lines = append(lines, lipgloss.NewStyle().Faint(true).Render(ruleLine(width)))
			}
		}
	}

	if len(lines) == 0 {
		return []string{lipgloss.NewStyle().Faint(true).Render("No items")}
	}
	if maxLines > 0 && len(lines) > maxLines {
		return lines[:maxLines]
	}
	return lines
}

func filterValues(values []string, query string) []string {
	if query == "" {
		return values
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		if strings.Contains(strings.ToLower(v), query) {
			out = append(out, v)
		}
	}
	return out
}

func filterPresets(values []FilterPreset, query string) []FilterPreset {
	if query == "" {
		return values
	}
	out := make([]FilterPreset, 0, len(values))
	for _, v := range values {
		pattern := strings.ToLower(v.Pattern)
		desc := strings.ToLower(v.Description)
		if strings.Contains(pattern, query) || strings.Contains(desc, query) {
			out = append(out, v)
		}
	}
	return out
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return string(r[:width])
	}
	return fmt.Sprintf("%s…", string(r[:width-1]))
}

func wrapPatternTwoLines(s string, width int) (string, string) {
	if width <= 0 {
		return "", ""
	}
	r := []rune(s)
	if len(r) <= width {
		return string(r), ""
	}
	first := string(r[:width])
	rest := r[width:]
	if len(rest) <= width {
		return first, string(rest)
	}
	if width <= 1 {
		return first, string(rest[:width])
	}
	return first, fmt.Sprintf("%s…", string(rest[:width-1]))
}

func ruleLine(width int) string {
	if width < 3 {
		return "---"
	}
	return strings.Repeat("-", width-1)
}

// --- Debugging --------------------------------------------------------------

func (d *FilterPalette) logState(reason string) {
	logging.Debugf("FilterPalette %s: visible=%t tab=%s focus=%s pCursor=%d hCursor=%d", reason, d.visible, d.activeTab, d.focusArea, d.presetCursor, d.historyCursor)
}
