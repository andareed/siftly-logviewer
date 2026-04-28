package siftly

import (
	"strings"
	"testing"
)

func TestNewModelFromCSVReaderPrecomputesTimeBounds(t *testing.T) {
	t.Parallel()

	csvData := strings.Join([]string{
		"time,details",
		"2026-04-22 10:00:00,alpha",
		"2026-04-22 11:30:00,beta",
	}, "\n")

	m, err := NewModelFromCSVReader(strings.NewReader(csvData), ColumnSchema{})
	if err != nil {
		t.Fatalf("NewModelFromCSVReader: %v", err)
	}

	if !m.table.derivedTimeData {
		t.Fatalf("derived time data should be precomputed")
	}
	if !m.table.hasTimeBounds {
		t.Fatalf("time bounds should be available")
	}
	if got := m.table.timeColumnIndex; got != 0 {
		t.Fatalf("time column index = %d want 0", got)
	}
	if got := len(m.table.rowTimes); got != 2 {
		t.Fatalf("rowTimes len = %d want 2", got)
	}
	if got := len(m.table.rowHasTimes); got != 2 {
		t.Fatalf("rowHasTimes len = %d want 2", got)
	}
	if !m.table.rowHasTimes[0] || !m.table.rowHasTimes[1] {
		t.Fatalf("rowHasTimes should be true for both rows")
	}
	if !m.table.timeMin.Before(m.table.timeMax) {
		t.Fatalf("time bounds should be ordered, got min=%v max=%v", m.table.timeMin, m.table.timeMax)
	}
}
