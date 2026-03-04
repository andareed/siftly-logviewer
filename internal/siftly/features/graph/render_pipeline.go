package graph

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
)

func collectPoints(in Input, maxKeys int) (order []string, series map[string][]point, minTS, maxTS int64, msg string) {
	series = make(map[string][]point, maxKeys)
	order = make([]string, 0, maxKeys)
	hasTS := false

	for _, idx := range in.FilteredIndices {
		if idx < 0 || idx >= len(in.Rows) {
			continue
		}
		row := in.Rows[idx]
		if in.TimeColumn >= len(row) || in.SeriesColumn >= len(row) || in.ValueColumn >= len(row) {
			continue
		}

		ts, err := strconv.ParseInt(strings.TrimSpace(row[in.TimeColumn]), 10, 64)
		if err != nil {
			continue
		}
		value, err := strconv.ParseFloat(strings.TrimSpace(row[in.ValueColumn]), 64)
		if err != nil {
			continue
		}

		key := strings.TrimSpace(row[in.SeriesColumn])
		if key == "" {
			key = "(empty)"
		}
		if _, exists := series[key]; !exists {
			if len(order) >= maxKeys {
				return nil, nil, 0, 0, fmt.Sprintf("Too much data, filter down to %d", maxKeys)
			}
			order = append(order, key)
			series[key] = nil
		}
		series[key] = append(series[key], point{ts: ts, val: value})

		if !hasTS {
			minTS = ts
			maxTS = ts
			hasTS = true
		} else {
			if ts < minTS {
				minTS = ts
			}
			if ts > maxTS {
				maxTS = ts
			}
		}
	}

	if len(order) == 0 {
		return nil, nil, 0, 0, "No numeric values"
	}
	if !hasTS {
		return nil, nil, 0, 0, "No timestamps"
	}

	for _, key := range order {
		pts := series[key]
		if len(pts) <= 1 {
			continue
		}
		sort.SliceStable(pts, func(i, j int) bool { return pts[i].ts < pts[j].ts })
		series[key] = pts
	}

	return order, series, minTS, maxTS, ""
}

func graphGeometry(totalHeight int) (plotHeight int, includeAxis bool, includeCursor bool, includeLegend bool) {
	switch {
	case totalHeight <= 2:
		return totalHeight, false, false, false
	case totalHeight == 3:
		return 1, true, true, false
	default:
		return totalHeight - 3, true, true, true
	}
}

func renderPlot(width int, height int, sampled [][]float64, palette []lipgloss.Color, scaleMode ScaleMode, layoutMode LayoutMode) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	if len(sampled) == 0 {
		return ui.RenderGraphMessage(width, height, "No numeric values")
	}

	if layoutMode == LayoutSplit && len(sampled) > 1 {
		if split := renderSplitPlot(width, height, sampled, palette, scaleMode); split != "" {
			return split
		}
	}
	return renderOverlayPlot(width, height, sampled, palette, scaleMode)
}

func renderSplitPlot(width int, height int, sampled [][]float64, palette []lipgloss.Color, scaleMode ScaleMode) string {
	heights := distributeRows(height, len(sampled))
	if len(heights) != len(sampled) {
		return ""
	}

	parts := make([]string, 0, len(sampled))
	for i := range sampled {
		colors := []lipgloss.Color{}
		if len(palette) > 0 {
			colors = append(colors, palette[i%len(palette)])
		}
		part := renderOverlayPlot(width, heights[i], [][]float64{sampled[i]}, colors, scaleMode)
		parts = append(parts, part)
	}
	return strings.Join(parts, "\n")
}

func distributeRows(total int, parts int) []int {
	if total <= 0 || parts <= 0 || total < parts {
		return nil
	}
	out := make([]int, parts)
	base := total / parts
	remainder := total % parts
	for i := 0; i < parts; i++ {
		h := base
		if i < remainder {
			h++
		}
		if h < 1 {
			h = 1
		}
		out[i] = h
	}
	return out
}
