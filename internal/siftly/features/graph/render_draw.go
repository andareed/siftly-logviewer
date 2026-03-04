package graph

import (
	"math"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
)

func renderOverlayPlot(width int, height int, sampled [][]float64, palette []lipgloss.Color, scaleMode ScaleMode) string {
	normalized, hasValues := normalizeSeries(sampled, scaleMode)
	if !hasValues {
		return ui.RenderGraphMessage(width, height, "No numeric values")
	}

	cells := make([][]uint8, height)
	colorGrid := make([][]int, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]uint8, width)
		colorGrid[y] = make([]int, width)
		for x := 0; x < width; x++ {
			colorGrid[y][x] = -1
		}
	}

	for sIdx, values := range normalized {
		if len(values) == 0 {
			continue
		}
		drawSeriesBraille(cells, colorGrid, values, width, height, sIdx)
	}

	return renderBrailleGrid(cells, colorGrid, palette)
}

func normalizeSeries(sampled [][]float64, scaleMode ScaleMode) ([][]float64, bool) {
	transformed := make([][]float64, len(sampled))
	globalMin := 0.0
	globalMax := 0.0
	hasGlobal := false

	for sIdx, values := range sampled {
		out := make([]float64, len(values))
		for i, v := range values {
			if math.IsNaN(v) {
				out[i] = math.NaN()
				continue
			}
			tv := transformValue(v, scaleMode)
			if math.IsNaN(tv) {
				out[i] = math.NaN()
				continue
			}
			out[i] = tv
			if !hasGlobal {
				globalMin = tv
				globalMax = tv
				hasGlobal = true
			} else {
				if tv < globalMin {
					globalMin = tv
				}
				if tv > globalMax {
					globalMax = tv
				}
			}
		}
		transformed[sIdx] = out
	}

	if !hasGlobal {
		return nil, false
	}

	out := make([][]float64, len(transformed))
	rangeV := globalMax - globalMin
	for sIdx, values := range transformed {
		norm := make([]float64, len(values))
		for i, v := range values {
			if math.IsNaN(v) {
				norm[i] = math.NaN()
				continue
			}
			if rangeV == 0 {
				norm[i] = 0.5
				continue
			}
			n := (v - globalMin) / rangeV
			if n < 0 {
				n = 0
			}
			if n > 1 {
				n = 1
			}
			norm[i] = n
		}
		out[sIdx] = norm
	}
	return out, true
}

func transformValue(v float64, mode ScaleMode) float64 {
	switch mode {
	case ScaleLog1P:
		if v < 0 {
			v = 0
		}
		return math.Log1p(v)
	case ScaleSymLog:
		if v == 0 {
			return 0
		}
		sign := 1.0
		if v < 0 {
			sign = -1.0
		}
		return sign * math.Log1p(math.Abs(v))
	default:
		return v
	}
}

func drawSeriesBraille(cells [][]uint8, colorGrid [][]int, normalized []float64, width int, height int, colorIdx int) {
	subHeight := height * 4
	if subHeight <= 0 {
		return
	}
	prevY := -1
	for x := 0; x < width; x++ {
		i := sampleIndex(len(normalized), width, x)
		n := normalized[i]
		if math.IsNaN(n) {
			continue
		}
		y := normalizedToSubY(n, subHeight)
		if prevY >= 0 {
			start := prevY
			end := y
			if start > end {
				start, end = end, start
			}
			for yfill := start; yfill <= end; yfill++ {
				setBraillePixel(cells, colorGrid, x, yfill, colorIdx)
			}
		}
		setBraillePixel(cells, colorGrid, x, y, colorIdx)
		prevY = y
	}
}

func normalizedToSubY(normalized float64, subHeight int) int {
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}
	y := int(math.Round((1 - normalized) * float64(subHeight-1)))
	if y < 0 {
		return 0
	}
	if y >= subHeight {
		return subHeight - 1
	}
	return y
}

func setBraillePixel(cells [][]uint8, colorGrid [][]int, x int, ySub int, colorIdx int) {
	if len(cells) == 0 || len(cells[0]) == 0 {
		return
	}
	height := len(cells)
	width := len(cells[0])
	if x < 0 || x >= width {
		return
	}
	if ySub < 0 {
		ySub = 0
	}
	maxSub := height*4 - 1
	if ySub > maxSub {
		ySub = maxSub
	}

	row := ySub / 4
	offset := ySub % 4
	var bit uint8
	switch offset {
	case 0:
		bit = 0x01
	case 1:
		bit = 0x02
	case 2:
		bit = 0x04
	default:
		bit = 0x40
	}

	cells[row][x] |= bit
	current := colorGrid[row][x]
	switch {
	case current == -1:
		colorGrid[row][x] = colorIdx
	case current == colorIdx:
		// no-op
	default:
		colorGrid[row][x] = -2
	}
}

func renderBrailleGrid(cells [][]uint8, colorGrid [][]int, colors []lipgloss.Color) string {
	height := len(cells)
	if height == 0 {
		return ""
	}
	width := len(cells[0])

	var styles []lipgloss.Style
	if len(colors) > 0 {
		styles = make([]lipgloss.Style, len(colors))
		for i, c := range colors {
			styles[i] = lipgloss.NewStyle().Foreground(c)
		}
	}

	var b strings.Builder
	for y := 0; y < height; y++ {
		if y > 0 {
			b.WriteByte('\n')
		}
		for x := 0; x < width; x++ {
			bits := cells[y][x]
			if bits == 0 {
				b.WriteByte(' ')
				continue
			}
			r := rune(0x2800) + rune(bits)
			cidx := colorGrid[y][x]
			if cidx >= 0 && cidx < len(styles) {
				b.WriteString(styles[cidx].Render(string(r)))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
