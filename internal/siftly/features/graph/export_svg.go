package graph

import (
	"bytes"
	"fmt"
	"html"
	"math"
	"time"
)

type ExportOptions struct {
	Width  int
	Height int
	Title  string
}

func ExportSVG(in Input, opts ExportOptions) ([]byte, error) {
	if opts.Width <= 0 {
		opts.Width = 1920
	}
	if opts.Height <= 0 {
		opts.Height = 1080
	}
	if opts.Width < 720 {
		opts.Width = 720
	}
	if opts.Height < 420 {
		opts.Height = 420
	}
	if in.TimeColumn < 0 || in.SeriesColumn < 0 || in.ValueColumn < 0 {
		return nil, fmt.Errorf("graph columns not configured")
	}

	maxKeys := in.MaxKeys
	if maxKeys <= 0 {
		maxKeys = 8
	}
	order, series, minTS, maxTS, msg := collectPoints(in, maxKeys)
	if msg != "" {
		return nil, fmt.Errorf("%s", msg)
	}
	if maxTS <= minTS {
		maxTS = minTS + 1
	}

	marginLeft := 92.0
	marginRight := 240.0
	marginTop := 92.0
	marginBottom := 96.0
	plotX := marginLeft
	plotY := marginTop
	plotW := float64(opts.Width) - marginLeft - marginRight
	plotH := float64(opts.Height) - marginTop - marginBottom
	if plotW < 240 || plotH < 180 {
		return nil, fmt.Errorf("graph export size too small")
	}

	sampleCount := int(math.Round(plotW))
	if sampleCount < 2 {
		sampleCount = 2
	}
	aggMode := NormalizeAggregateMode(in.AggregateMode)
	fillMode := NormalizeFillMode(in.FillMode)
	scaleMode := NormalizeScaleMode(in.ScaleMode)
	sampled := make([][]float64, 0, len(order))
	for _, key := range order {
		sampled = append(sampled, sampleSeriesByTime(series[key], minTS, maxTS, sampleCount, aggMode, fillMode))
	}

	rawMin, rawMax, scaledMin, scaledMax, ok := exportRanges(sampled, scaleMode)
	if !ok {
		return nil, fmt.Errorf("no numeric values")
	}

	title := opts.Title
	if title == "" {
		title = "Siftly graph export"
	}

	palette := []string{
		"#e15759",
		"#00a6a6",
		"#f2b134",
		"#4e79d9",
		"#8f63d7",
		"#f28e2b",
		"#59a14f",
		"#b07aa1",
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`+"\n", opts.Width, opts.Height, opts.Width, opts.Height)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="#fbfaf7"/>`+"\n")
	fmt.Fprintf(&b, `<text x="40" y="44" font-family="ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif" font-size="24" font-weight="700" fill="#202124">%s</text>`+"\n", svgText(title))
	fmt.Fprintf(&b, `<text x="40" y="70" font-family="ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif" font-size="13" fill="#62666d">filtered rows: %d | series: %d | scale: %s | aggregate: %s</text>`+"\n", len(in.FilteredIndices), len(order), svgText(string(scaleMode)), svgText(string(aggMode)))

	writeGridAndAxes(&b, plotX, plotY, plotW, plotH, rawMin, rawMax, scaledMin, scaledMax, scaleMode, minTS, maxTS)

	for sIdx, values := range sampled {
		color := palette[sIdx%len(palette)]
		writeSeriesPaths(&b, values, plotX, plotY, plotW, plotH, scaledMin, scaledMax, scaleMode, color)
	}

	writeLegend(&b, order, palette, plotX+plotW+34, plotY+8)
	fmt.Fprintf(&b, "</svg>\n")
	return b.Bytes(), nil
}

func exportRanges(sampled [][]float64, scaleMode ScaleMode) (rawMin, rawMax, scaledMin, scaledMax float64, ok bool) {
	for _, values := range sampled {
		for _, raw := range values {
			if math.IsNaN(raw) {
				continue
			}
			scaled := transformValue(raw, scaleMode)
			if math.IsNaN(scaled) {
				continue
			}
			if !ok {
				rawMin, rawMax = raw, raw
				scaledMin, scaledMax = scaled, scaled
				ok = true
				continue
			}
			if raw < rawMin {
				rawMin = raw
			}
			if raw > rawMax {
				rawMax = raw
			}
			if scaled < scaledMin {
				scaledMin = scaled
			}
			if scaled > scaledMax {
				scaledMax = scaled
			}
		}
	}
	if ok && scaledMax == scaledMin {
		scaledMin -= 1
		scaledMax += 1
	}
	return rawMin, rawMax, scaledMin, scaledMax, ok
}

func writeGridAndAxes(b *bytes.Buffer, plotX, plotY, plotW, plotH float64, rawMin, rawMax, scaledMin, scaledMax float64, scaleMode ScaleMode, minTS, maxTS int64) {
	fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#ffffff" stroke="#d8d5ce" stroke-width="1"/>`+"\n", plotX, plotY, plotW, plotH)

	ticks := 5
	for i := 0; i <= ticks; i++ {
		ratio := float64(i) / float64(ticks)
		y := plotY + plotH - ratio*plotH
		scaled := scaledMin + ratio*(scaledMax-scaledMin)
		raw := inverseTransformValue(scaled, scaleMode)
		if scaleMode == ScaleLog1P && rawMin < 0 {
			raw = rawMin + ratio*(rawMax-rawMin)
		}
		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#ebe8e1" stroke-width="1"/>`+"\n", plotX, y, plotX+plotW, y)
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="end" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#62666d">%s</text>`+"\n", plotX-12, y+4, svgText(formatExportValue(raw)))
	}

	for i := 0; i <= 4; i++ {
		ratio := float64(i) / 4.0
		x := plotX + ratio*plotW
		ts := minTS + int64(ratio*float64(maxTS-minTS))
		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#f1eee8" stroke-width="1"/>`+"\n", x, plotY, x, plotY+plotH)
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="12" fill="#62666d">%s</text>`+"\n", x, plotY+plotH+28, svgText(formatExportTime(ts)))
	}

	fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#2f3337" stroke-width="1.25"/>`+"\n", plotX, plotY+plotH, plotX+plotW, plotY+plotH)
	fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#2f3337" stroke-width="1.25"/>`+"\n", plotX, plotY, plotX, plotY+plotH)
}

func writeSeriesPaths(b *bytes.Buffer, values []float64, plotX, plotY, plotW, plotH, scaledMin, scaledMax float64, scaleMode ScaleMode, color string) {
	if len(values) == 0 {
		return
	}
	rangeV := scaledMax - scaledMin
	if rangeV == 0 {
		rangeV = 1
	}

	var path bytes.Buffer
	drawing := false
	wrote := false
	for i, raw := range values {
		if math.IsNaN(raw) {
			drawing = false
			continue
		}
		scaled := transformValue(raw, scaleMode)
		if math.IsNaN(scaled) {
			drawing = false
			continue
		}
		x := plotX
		if len(values) > 1 {
			x += float64(i) * plotW / float64(len(values)-1)
		}
		n := (scaled - scaledMin) / rangeV
		if n < 0 {
			n = 0
		}
		if n > 1 {
			n = 1
		}
		y := plotY + (1-n)*plotH
		if !drawing {
			fmt.Fprintf(&path, "M %.2f %.2f ", x, y)
			drawing = true
		} else {
			fmt.Fprintf(&path, "L %.2f %.2f ", x, y)
		}
		wrote = true
	}
	if !wrote {
		return
	}
	fmt.Fprintf(b, `<path d="%s" fill="none" stroke="%s" stroke-width="2.2" stroke-linejoin="round" stroke-linecap="round" opacity="0.92"/>`+"\n", path.String(), color)
}

func writeLegend(b *bytes.Buffer, order []string, palette []string, x, y float64) {
	fmt.Fprintf(b, `<text x="%.1f" y="%.1f" font-family="ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif" font-size="13" font-weight="700" fill="#202124">Series</text>`+"\n", x, y)
	for i, key := range order {
		yy := y + 28 + float64(i*26)
		color := palette[i%len(palette)]
		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="3" stroke-linecap="round"/>`+"\n", x, yy-4, x+22, yy-4, color)
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" font-family="ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif" font-size="12" fill="#3d4147">%s</text>`+"\n", x+32, yy, svgText(truncateExportLabel(key, 28)))
	}
}

func inverseTransformValue(v float64, mode ScaleMode) float64 {
	switch mode {
	case ScaleLog1P:
		return math.Expm1(v)
	case ScaleSymLog:
		if v == 0 {
			return 0
		}
		sign := 1.0
		if v < 0 {
			sign = -1.0
		}
		return sign * math.Expm1(math.Abs(v))
	default:
		return v
	}
}

func formatExportValue(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= 1000:
		return fmt.Sprintf("%.0f", v)
	case abs >= 10:
		return fmt.Sprintf("%.1f", v)
	default:
		return fmt.Sprintf("%.2f", v)
	}
}

func formatExportTime(ts int64) string {
	return time.Unix(ts, 0).Format("01-02 15:04")
}

func truncateExportLabel(s string, max int) string {
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	if max <= 3 {
		return string(rs[:max])
	}
	return string(rs[:max-3]) + "..."
}

func svgText(s string) string {
	return html.EscapeString(s)
}
