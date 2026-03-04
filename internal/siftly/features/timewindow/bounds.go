package timewindow

import (
	"strconv"
	"strings"
	"time"
)

const (
	logTimeLayout  = "Mon Jan 02 15:04:05 MST 2006"
	dateTimeLayout = "2006-01-02 15:04:05"
)

type Bounds struct {
	Has             bool
	Min             time.Time
	Max             time.Time
	TimeColumnIndex int
	RowTimes        []time.Time
	RowHasTimes     []bool
}

func ComputeBounds(columnNames []string, rows [][]string) Bounds {
	timeCol := FindTimeColumnIndex(columnNames)
	b := Bounds{
		TimeColumnIndex: timeCol,
		RowTimes:        make([]time.Time, len(rows)),
		RowHasTimes:     make([]bool, len(rows)),
	}
	if timeCol < 0 {
		return b
	}

	hasAny := false
	var minTime time.Time
	var maxTime time.Time
	for i, row := range rows {
		if timeCol >= len(row) {
			continue
		}
		ts, ok := ParseLogTimestamp(row[timeCol])
		if !ok {
			continue
		}
		b.RowTimes[i] = ts
		b.RowHasTimes[i] = true
		if !hasAny {
			minTime = ts
			maxTime = ts
			hasAny = true
			continue
		}
		if ts.Before(minTime) {
			minTime = ts
		}
		if ts.After(maxTime) {
			maxTime = ts
		}
	}

	b.Has = hasAny
	b.Min = minTime
	b.Max = maxTime
	return b
}

func FindTimeColumnIndex(columnNames []string) int {
	preferred := map[string]struct{}{
		"time":      {},
		"timestamp": {},
	}
	fallback := map[string]struct{}{
		"date": {},
		"ts":   {},
	}

	fallbackIdx := -1
	for i, name := range columnNames {
		n := strings.TrimSpace(name)
		n = strings.TrimPrefix(n, "\ufeff")
		n = strings.ToLower(n)
		if _, ok := preferred[n]; ok {
			return i
		}
		if fallbackIdx < 0 {
			if _, ok := fallback[n]; ok {
				fallbackIdx = i
			}
		}
	}
	return fallbackIdx
}

func ParseLogTimestamp(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	if ts, ok := parseUnixTimestamp(raw); ok {
		return ts, true
	}

	if ts, err := time.Parse(logTimeLayout, raw); err == nil {
		return ts, true
	}
	if ts, err := time.Parse(dateTimeLayout, raw); err == nil {
		return ts, true
	}
	if ts, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return ts, true
	}

	if idx := strings.LastIndex(raw, ":"); idx != -1 {
		raw = strings.TrimSpace(raw[:idx])
	}
	if ts, err := time.Parse(logTimeLayout, raw); err == nil {
		return ts, true
	}
	if ts, err := time.Parse(dateTimeLayout, raw); err == nil {
		return ts, true
	}
	return time.Time{}, false
}

func parseUnixTimestamp(raw string) (time.Time, bool) {
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, false
	}

	abs := v
	if abs < 0 {
		abs = -abs
	}

	switch {
	case abs >= 100000000000000000:
		return time.Unix(0, v), true
	case abs >= 100000000000000:
		return time.UnixMicro(v), true
	case abs >= 100000000000:
		return time.UnixMilli(v), true
	default:
		return time.Unix(v, 0), true
	}
}

func Clamp(t, min, max time.Time) time.Time {
	if t.Before(min) {
		return min
	}
	if t.After(max) {
		return max
	}
	return t
}

func DefaultBounds(min, max time.Time) (time.Time, time.Time) {
	return min, max
}

func CursorTimestamp(filteredIndices []int, cursor int, rowHasTimes []bool, rowTimes []time.Time) (time.Time, bool) {
	if cursor < 0 || cursor >= len(filteredIndices) {
		return time.Time{}, false
	}
	rowIdx := filteredIndices[cursor]
	if rowIdx < 0 || rowIdx >= len(rowHasTimes) {
		return time.Time{}, false
	}
	if !rowHasTimes[rowIdx] {
		return time.Time{}, false
	}
	return rowTimes[rowIdx], true
}

func SetEdge(current Window, ts, min, max time.Time, setStart bool) Window {
	start := current.Start
	end := current.End
	if !current.Enabled || start.IsZero() || end.IsZero() {
		start = min
		end = max
	}
	if setStart {
		start = ts
	} else {
		end = ts
	}

	start = Clamp(start, min, max)
	end = Clamp(end, min, max)
	if start.After(end) {
		if setStart {
			end = start
		} else {
			start = end
		}
	}

	return Window{
		Enabled: true,
		Start:   start,
		End:     end,
	}
}
