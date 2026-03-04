package graph

import (
	"math"
	"strconv"
	"strings"
	"time"
)

func cursorTimestamp(in Input) int64 {
	if in.Cursor < 0 || in.Cursor >= len(in.FilteredIndices) {
		return 0
	}
	rowIdx := in.FilteredIndices[in.Cursor]
	if rowIdx < 0 || rowIdx >= len(in.Rows) {
		return 0
	}
	row := in.Rows[rowIdx]
	if in.TimeColumn < 0 || in.TimeColumn >= len(row) {
		return 0
	}
	ts, err := strconv.ParseInt(strings.TrimSpace(row[in.TimeColumn]), 10, 64)
	if err != nil {
		return 0
	}
	return ts
}

func renderTimeAxisLabels(width int, minTS, maxTS int64) string {
	if width <= 0 {
		return ""
	}
	line := make([]rune, width)
	for i := range line {
		line[i] = ' '
	}
	minLabel := formatTimestamp(minTS)
	maxLabel := formatTimestamp(maxTS)
	if len([]rune(minLabel))+len([]rune(maxLabel))+1 > width {
		minLabel = time.Unix(minTS, 0).Format("01-02 15:04")
		maxLabel = time.Unix(maxTS, 0).Format("01-02 15:04")
		if len([]rune(minLabel))+len([]rune(maxLabel))+1 > width {
			minLabel = time.Unix(minTS, 0).Format("15:04")
			maxLabel = time.Unix(maxTS, 0).Format("15:04")
		}
	}

	placeLabel(line, 0, minLabel)
	placeLabel(line, width-len([]rune(maxLabel)), maxLabel)

	return string(line)
}

func renderTimeCursorLine(width int, minTS, maxTS, cursorTS int64) string {
	if width <= 0 {
		return ""
	}
	line := make([]rune, width)
	for i := range line {
		line[i] = '─'
	}
	if maxTS <= minTS {
		maxTS = minTS + 1
	}
	if cursorTS < minTS {
		cursorTS = minTS
	}
	if cursorTS > maxTS {
		cursorTS = maxTS
	}
	pos := 0
	if width > 1 {
		ratio := float64(cursorTS-minTS) / float64(maxTS-minTS) * float64(width-1)
		pos = int(math.Ceil(ratio))
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= width {
		pos = width - 1
	}
	line[pos] = '▲'
	return string(line)
}

func formatTimestamp(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func placeLabel(line []rune, start int, label string) {
	if start < 0 || start >= len(line) {
		return
	}
	rs := []rune(label)
	for i := 0; i < len(rs); i++ {
		pos := start + i
		if pos < 0 || pos >= len(line) {
			return
		}
		line[pos] = rs[i]
	}
}
