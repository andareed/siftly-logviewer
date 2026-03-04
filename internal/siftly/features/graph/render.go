package graph

import (
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
)

type Input struct {
	Width           int
	Height          int
	Rows            [][]string
	FilteredIndices []int
	Cursor          int
	TimeColumn      int
	SeriesColumn    int
	ValueColumn     int
	MaxKeys         int
	ScaleMode       string
	AggregateMode   string
	LayoutMode      string
	FillMode        string
}

type point struct {
	ts  int64
	val float64
}

// Prepared stores the static graph output and metadata needed to draw
// the cursor indicator without rebuilding the whole graph.
type Prepared struct {
	width        int
	static       string
	beforeCursor string
	afterCursor  string
	minTS        int64
	maxTS        int64
	withCursor   bool
}

// Render draws the graph in one pass.
func Render(in Input) string {
	prepared := Prepare(in)
	return prepared.Render(cursorTimestamp(in))
}

// Prepare builds graph data that is stable across cursor movement.
func Prepare(in Input) Prepared {
	if in.Width <= 0 || in.Height <= 0 {
		return Prepared{width: in.Width}
	}
	if in.TimeColumn < 0 || in.SeriesColumn < 0 || in.ValueColumn < 0 {
		return Prepared{
			width:  in.Width,
			static: ui.RenderGraphMessage(in.Width, in.Height, "Graph columns not configured"),
		}
	}

	maxKeys := in.MaxKeys
	if maxKeys <= 0 {
		maxKeys = 5
	}

	order, series, minTS, maxTS, msg := collectPoints(in, maxKeys)
	if msg != "" {
		return Prepared{
			width:  in.Width,
			static: ui.RenderGraphMessage(in.Width, in.Height, msg),
		}
	}

	aggMode := NormalizeAggregateMode(in.AggregateMode)
	fillMode := NormalizeFillMode(in.FillMode)
	sampled := make([][]float64, 0, len(order))
	for _, key := range order {
		sampled = append(sampled, sampleSeriesByTime(series[key], minTS, maxTS, in.Width, aggMode, fillMode))
	}

	scaleMode := NormalizeScaleMode(in.ScaleMode)
	layoutMode := NormalizeLayoutMode(in.LayoutMode)
	palette := []lipgloss.Color{
		lipgloss.Color("#ff6b6b"),
		lipgloss.Color("#4ecdc4"),
		lipgloss.Color("#ffe66d"),
		lipgloss.Color("#5b8def"),
		lipgloss.Color("#a66dd4"),
		lipgloss.Color("#f08a24"),
	}

	plotHeight, includeAxis, includeCursor, includeLegend := graphGeometry(in.Height)
	if plotHeight <= 0 {
		return Prepared{
			width:  in.Width,
			static: ui.RenderGraphMessage(in.Width, in.Height, "Graph height too small"),
		}
	}

	plot := renderPlot(in.Width, plotHeight, sampled, palette, scaleMode, layoutMode)
	axisLabels := ""
	if includeAxis {
		axisLabels = renderTimeAxisLabels(in.Width, minTS, maxTS)
	}
	legend := ""
	if includeLegend {
		legend = ui.RenderLegendLine(in.Width, order, palette)
	}

	if !includeCursor {
		content := plot
		if axisLabels != "" {
			content += "\n" + axisLabels
		}
		if legend != "" {
			content += "\n" + legend
		}
		return Prepared{width: in.Width, static: content}
	}

	before := plot
	if axisLabels != "" {
		before += "\n" + axisLabels
	}
	after := ""
	if legend != "" {
		after = "\n" + legend
	}
	return Prepared{
		width:        in.Width,
		beforeCursor: before,
		afterCursor:  after,
		minTS:        minTS,
		maxTS:        maxTS,
		withCursor:   true,
	}
}

// Render draws the final graph for the current cursor timestamp.
func (p Prepared) Render(cursorTS int64) string {
	if !p.withCursor {
		return p.static
	}
	cursorLine := renderTimeCursorLine(p.width, p.minTS, p.maxTS, cursorTS)
	if p.afterCursor == "" {
		return p.beforeCursor + "\n" + cursorLine
	}
	return p.beforeCursor + "\n" + cursorLine + p.afterCursor
}
