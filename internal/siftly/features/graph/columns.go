package graph

import "strings"

func ResolveColumnIndices(header []string, cfg Config) (timeCol int, seriesCol int, valueCol int, ok bool) {
	timeCol = indexByName(header, cfg.TimeColumn)
	seriesCol = indexByName(header, cfg.SeriesColumn)
	valueCol = indexByName(header, cfg.ValueColumn)
	ok = timeCol >= 0 && seriesCol >= 0 && valueCol >= 0
	return timeCol, seriesCol, valueCol, ok
}

func indexByName(header []string, name string) int {
	target := strings.ToLower(strings.TrimSpace(name))
	if target == "" {
		return -1
	}
	for i, col := range header {
		if strings.ToLower(strings.TrimSpace(col)) == target {
			return i
		}
	}
	return -1
}
