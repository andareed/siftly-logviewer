package timewindow

import (
	"strconv"
	"strings"
	"time"
)

const (
	logTimeLayout     = "Mon Jan 02 15:04:05 2006"
	dateTimeLayout    = "2006-01-02 15:04:05"
	secondsPerHour    = 60 * 60
	secondsPerMinute  = 60
	unknownZoneOffset = 0
)

var knownZoneOffsets = map[string]int{
	"GMT":  0,
	"UTC":  0,
	"BST":  1 * secondsPerHour,
	"CET":  1 * secondsPerHour,
	"CEST": 2 * secondsPerHour,
	"WET":  0,
	"WEST": 1 * secondsPerHour,
	"EET":  2 * secondsPerHour,
	"EEST": 3 * secondsPerHour,
	"CAT":  2 * secondsPerHour,
	"SAST": 2 * secondsPerHour,
	"EAT":  3 * secondsPerHour,
}

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

	if ts, ok := parseHostlogTimestamp(raw); ok {
		return ts, true
	}
	if ts, err := time.Parse(dateTimeLayout, raw); err == nil {
		return ts, true
	}
	if ts, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return ts, true
	}
	return time.Time{}, false
}

func parseHostlogTimestamp(raw string) (time.Time, bool) {
	fields := strings.Fields(raw)
	if len(fields) != 6 {
		return time.Time{}, false
	}

	year := fields[5]
	if idx := strings.Index(year, ":"); idx != -1 {
		if !isSignedDecimal(year[idx+1:]) {
			return time.Time{}, false
		}
		year = year[:idx]
	}

	zoneName, zoneOffset, ok := parseZoneToken(fields[4])
	if !ok {
		return time.Time{}, false
	}

	rawNoZone := strings.Join([]string{
		fields[0],
		fields[1],
		fields[2],
		fields[3],
		year,
	}, " ")
	ts, err := time.ParseInLocation(logTimeLayout, rawNoZone, time.FixedZone(zoneName, zoneOffset))
	if err != nil {
		return time.Time{}, false
	}
	return ts, true
}

func parseZoneToken(raw string) (string, int, bool) {
	zone := strings.ToUpper(strings.TrimSpace(raw))
	if zone == "" {
		return "", 0, false
	}

	for _, prefix := range []string{"GMT", "UTC"} {
		if zone == prefix {
			return prefix, 0, true
		}
		if strings.HasPrefix(zone, prefix) {
			offset, ok := parseNumericOffset(zone[len(prefix):])
			if ok {
				return zone, offset, true
			}
		}
	}

	if offset, ok := parseNumericOffset(zone); ok {
		return zone, offset, true
	}
	if offset, ok := knownZoneOffsets[zone]; ok {
		return zone, offset, true
	}
	if isZoneAbbreviation(zone) {
		// Unknown abbreviations are accepted so rows stay usable; exact conversion needs a configured offset.
		return zone, unknownZoneOffset, true
	}
	return "", 0, false
}

func parseNumericOffset(raw string) (int, bool) {
	if len(raw) < 2 {
		return 0, false
	}

	sign := 1
	switch raw[0] {
	case '+':
	case '-':
		sign = -1
	default:
		return 0, false
	}

	body := raw[1:]
	var hourStr string
	var minuteStr string
	if strings.Contains(body, ":") {
		parts := strings.Split(body, ":")
		if len(parts) != 2 {
			return 0, false
		}
		hourStr = parts[0]
		minuteStr = parts[1]
	} else {
		switch len(body) {
		case 1, 2:
			hourStr = body
			minuteStr = "0"
		case 3:
			hourStr = body[:1]
			minuteStr = body[1:]
		case 4:
			hourStr = body[:2]
			minuteStr = body[2:]
		default:
			return 0, false
		}
	}

	hours, err := strconv.Atoi(hourStr)
	if err != nil {
		return 0, false
	}
	minutes, err := strconv.Atoi(minuteStr)
	if err != nil {
		return 0, false
	}
	if hours > 23 || minutes > 59 {
		return 0, false
	}
	return sign * ((hours * secondsPerHour) + (minutes * secondsPerMinute)), true
}

func isZoneAbbreviation(zone string) bool {
	if len(zone) < 2 || len(zone) > 8 {
		return false
	}
	for _, r := range zone {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func isSignedDecimal(raw string) bool {
	if raw == "" {
		return false
	}
	if raw[0] == '+' || raw[0] == '-' {
		raw = raw[1:]
	}
	if raw == "" {
		return false
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
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
