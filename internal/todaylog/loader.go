package todaylog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly"
)

var todaylogHeader = []string{"date", "timestamp", "pid", "process", "key", "value"}

func LoadModelAuto(path string) (*siftly.Model, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	default:
		return initiateModelFromStats(path)
	}
}

func initiateModelFromStats(statsFile string) (*siftly.Model, error) {
	defer logging.TimeOperation("load todaylog stats")()

	file, err := os.Open(statsFile)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", statsFile, err)
	}
	defer file.Close()

	builder, err := siftly.NewModelBuilder(todaylogHeader, todaylogColumnSchema())
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	profile := logging.IsDebugMode()
	var parseDuration time.Duration
	var addRowDuration time.Duration
	var buildDuration time.Duration

	lineNo := 0
	rowCount := 0
	skippedCount := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parseStart := time.Time{}
		if profile {
			parseStart = time.Now()
		}
		record, fieldCount, err := parseStatsLine(line)
		if profile {
			parseDuration += time.Since(parseStart)
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if fieldCount < 6 {
			skippedCount++
			logging.Warnf("Skipping line %d: expected at least 6 fields, got %d", lineNo, fieldCount)
			continue
		}

		addStart := time.Time{}
		if profile {
			addStart = time.Now()
		}
		builder.AddRowOwned(record, lineNo)
		if profile {
			addRowDuration += time.Since(addStart)
		}
		rowCount++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %q: %w", statsFile, err)
	}

	buildStart := time.Time{}
	if profile {
		buildStart = time.Now()
	}
	m := builder.Build()
	if profile {
		buildDuration += time.Since(buildStart)
		logging.Infof(
			"todaylog loader phases: rows=%d skipped=%d parse=%s addRow=%s build=%s",
			rowCount,
			skippedCount,
			parseDuration.Round(time.Millisecond),
			addRowDuration.Round(time.Millisecond),
			buildDuration.Round(time.Millisecond),
		)
	}
	m.InitialPath = statsFile
	m.SetStyles(SiftlyStyles())
	m.SetGraphConfig(siftly.GraphConfig{
		Enabled:      true,
		TimeColumn:   "timestamp",
		SeriesColumn: "key",
		ValueColumn:  "value",
		Height:       16,
		MaxKeys:      8,
		ScaleMode:    "log1p",
		Aggregate:    "last",
		Layout:       "overlay",
		FillMode:     "none",
	})
	m.InitialiseView()
	return m, nil
}

func parseStatsLine(line string) ([]string, int, error) {
	fields := make([]string, 0, 6)
	rest := line
	fieldCount := 0

	for fieldCount < 5 {
		token, next, ok := cutNextField(rest)
		if !ok {
			return nil, fieldCount, nil
		}
		fields = append(fields, token)
		fieldCount++
		rest = next
	}

	value := strings.TrimSpace(rest)
	if value == "" {
		return nil, fieldCount, nil
	}
	fieldCount++

	timestamp, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, fieldCount, fmt.Errorf("parse timestamp %q: %w", fields[1], err)
	}
	if !isDecimal(fields[3]) {
		return nil, fieldCount, fmt.Errorf("parse pid %q: invalid decimal", fields[3])
	}

	return []string{
		time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"),
		fields[1],
		fields[3],
		fields[2],
		fields[4],
		value,
	}, fieldCount, nil
}

func cutNextField(s string) (field string, rest string, ok bool) {
	i := 0
	for i < len(s) && isSpace(s[i]) {
		i++
	}
	if i >= len(s) {
		return "", "", false
	}

	start := i
	for i < len(s) && !isSpace(s[i]) {
		i++
	}

	field = s[start:i]
	for i < len(s) && isSpace(s[i]) {
		i++
	}
	return field, s[i:], true
}

func isSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
}

func isDecimal(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
