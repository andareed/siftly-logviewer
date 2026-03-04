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

func LoadModelAuto(path string) (*siftly.Model, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	default:
		return initiateModelFromStats(path)
	}
}

func initiateModelFromStats(statsFile string) (*siftly.Model, error) {
	rawStats, err := loadStatsFile(statsFile)
	if err != nil {
		return nil, err
	}
	records, err := parseStatsRows(rawStats)
	if err != nil {
		return nil, err
	}

	m, err := siftly.NewModelFromRecords(records, todaylogColumnSchema())
	if err != nil {
		return nil, err
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
		// Keep sparse hourly/daily series as gaps instead of drawing held lines.
		FillMode: "none",
	})
	m.InitialiseView()
	return m, nil
}

func loadStatsFile(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()

	var rows [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rows = append(rows, strings.Fields(line))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	return rows, nil
}

func parseStatsRows(rows [][]string) ([][]string, error) {
	records := [][]string{
		{"date", "timestamp", "pid", "process", "key", "value"},
	}

	for i, row := range rows {
		if len(row) < 6 {
			logging.Warnf("Skipping line %d: expected at least 6 fields, got %d", i+1, len(row))
			continue
		}

		timestamp, err := strconv.ParseInt(row[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: parse timestamp %q: %w", i+1, row[1], err)
		}

		pid, err := strconv.Atoi(row[3])
		if err != nil {
			return nil, fmt.Errorf("line %d: parse pid %q: %w", i+1, row[3], err)
		}

		record := []string{
			time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"),
			strconv.FormatInt(timestamp, 10),
			strconv.Itoa(pid),
			row[2],
			row[4],
			strings.Join(row[5:], " "),
		}
		records = append(records, record)
	}

	return records, nil
}
