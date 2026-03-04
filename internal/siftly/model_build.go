package siftly

import (
	"fmt"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

// ColumnSchema defines how column names map to roles and how layout defaults are applied.
// Resolution order for min-width/weight:
// 1) DefaultMinWidth/DefaultWeight
// 2) RoleDefaults[role]
// 3) ColumnDefaults[normalized-column-name]
type ColumnSchema struct {
	DefaultMinWidth int
	DefaultWeight   float64
	RoleForName     func(name string) ui.ColumnRole
	RoleDefaults    map[ui.ColumnRole]RoleLayout
	ColumnDefaults  map[string]RoleLayout
}

type RoleLayout struct {
	MinWidth int
	Weight   float64
}

// NewModelFromRecords builds a model from CSV records (including header row).
func NewModelFromRecords(records [][]string, schema ColumnSchema) (*Model, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records provided")
	}
	if len(records[0]) == 0 {
		return nil, fmt.Errorf("header row is empty")
	}

	roleFn := schema.RoleForName
	if roleFn == nil {
		roleFn = func(string) ui.ColumnRole { return ui.RoleNormal }
	}
	defMin := schema.DefaultMinWidth
	if defMin == 0 {
		defMin = 8
	}
	defWeight := schema.DefaultWeight
	if defWeight == 0 {
		defWeight = 1.0
	}

	rawHeader := records[0]
	cols := make([]ui.ColumnMeta, len(rawHeader))
	for i, name := range rawHeader {
		role := roleFn(name)
		minWidth := defMin
		weight := defWeight

		if layout, ok := schema.RoleDefaults[role]; ok {
			if layout.MinWidth > 0 {
				minWidth = layout.MinWidth
			}
			if layout.Weight > 0 {
				weight = layout.Weight
			}
		}

		colName := strings.ToLower(strings.TrimSpace(name))
		if layout, ok := schema.ColumnDefaults[colName]; ok {
			if layout.MinWidth > 0 {
				minWidth = layout.MinWidth
			}
			if layout.Weight > 0 {
				weight = layout.Weight
			}
		}

		cols[i] = ui.ColumnMeta{
			Name:     name,
			Index:    i,
			Role:     role,
			Visible:  true,
			MinWidth: minWidth,
			Weight:   weight,
		}
	}

	rows := make([]Row, 0, len(records)-1)
	hasData := make([]bool, len(cols)) // track which columns ever have content

	for i, csvRow := range records[1:] {
		for colIdx := range cols {
			if colIdx < len(csvRow) && strings.TrimSpace(csvRow[colIdx]) != "" {
				hasData[colIdx] = true
			}
		}

		row := Row{
			Cols: csvRow,
		}
		row.ID = row.ComputeID()
		row.OriginalIndex = i + 1
		rows = append(rows, row)
	}

	for i := range cols {
		if !hasData[i] {
			if cols[i].Role != ui.RolePrimary {
				cols[i].Visible = false
			}
			cols[i].Weight = 0
			cols[i].Width = 0
		}
	}

	return &Model{
		table: tableState{
			header:      cols,
			rows:        rows,
			markedRows:  make(map[uint64]ui.MarkColor),
			commentRows: make(map[uint64]string),
			sortColumn:  -1,
		},
		view: viewState{mode: modeView},
	}, nil
}
