package siftly

import (
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
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

func TestColumnSchemaAppliesAndRefreshesWrappedLineBudget(t *testing.T) {
	schema := ColumnSchema{
		RoleForName: func(name string) ui.ColumnRole {
			if strings.EqualFold(name, "details") {
				return ui.RolePrimary
			}
			return ui.RoleNormal
		},
		RoleDefaults: map[ui.ColumnRole]RoleLayout{
			ui.RolePrimary: {MinWidth: 24, Weight: 5, WrapLines: 4},
		},
	}
	m, err := NewModelFromRecords([][]string{{"host", "details"}, {"one", "long value"}}, schema)
	if err != nil {
		t.Fatalf("build model: %v", err)
	}
	if got := m.table.header[1].WrapLines; got != 4 {
		t.Fatalf("details WrapLines=%d want 4", got)
	}

	m.table.header = []ui.ColumnMeta{m.table.header[1], m.table.header[0]}
	m.table.header[0].Visible = false
	m.table.header[0].MinWidth = 37
	m.table.header[0].WrapLines = 0
	m.ApplyColumnSchema(schema)
	got := m.table.header[0]
	if got.Role != ui.RolePrimary || got.WrapLines != 4 {
		t.Fatalf("schema semantics not refreshed: %+v", got)
	}
	if got.Visible || got.MinWidth != 37 || got.Index != 1 {
		t.Fatalf("schema refresh changed user layout: %+v", got)
	}
}
