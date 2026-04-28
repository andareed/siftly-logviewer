package siftly

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
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

type modelBuilder struct {
	header          []ui.ColumnMeta
	hasData         []bool
	rows            []Row
	timeColumnIndex int
	rowTimes        []time.Time
	rowHasTimes     []bool
	hasTimeBounds   bool
	timeMin         time.Time
	timeMax         time.Time
}

type ModelBuilder struct {
	builder *modelBuilder
}

// NewModelFromRecords builds a model from CSV records (including header row).
func NewModelFromRecords(records [][]string, schema ColumnSchema) (*Model, error) {
	defer logging.TimeOperation("build model from records")()

	if len(records) == 0 {
		return nil, fmt.Errorf("no records provided")
	}

	builder, err := newModelBuilder(records[0], schema)
	if err != nil {
		return nil, err
	}

	for i, record := range records[1:] {
		builder.addRow(record, i+1, true)
	}

	return builder.finish(), nil
}

func NewModelFromCSVReader(r io.Reader, schema ColumnSchema) (*Model, error) {
	defer logging.TimeOperation("build model from csv")()

	reader := csv.NewReader(r)
	reader.ReuseRecord = true

	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("no records provided")
		}
		return nil, fmt.Errorf("read header: %w", err)
	}

	builder, err := newModelBuilder(header, schema)
	if err != nil {
		return nil, err
	}

	originalIndex := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", originalIndex, err)
		}
		builder.addRow(record, originalIndex, true)
		originalIndex++
	}

	return builder.finish(), nil
}

func NewModelBuilder(header []string, schema ColumnSchema) (*ModelBuilder, error) {
	builder, err := newModelBuilder(header, schema)
	if err != nil {
		return nil, err
	}
	return &ModelBuilder{builder: builder}, nil
}

func (b *ModelBuilder) AddRow(cols []string, originalIndex int) {
	if b == nil || b.builder == nil {
		return
	}
	b.builder.addRow(cols, originalIndex, true)
}

func (b *ModelBuilder) AddRowOwned(cols []string, originalIndex int) {
	if b == nil || b.builder == nil {
		return
	}
	b.builder.addRow(cols, originalIndex, false)
}

func (b *ModelBuilder) Build() *Model {
	if b == nil || b.builder == nil {
		return nil
	}
	return b.builder.finish()
}

func newModelBuilder(rawHeader []string, schema ColumnSchema) (*modelBuilder, error) {
	if len(rawHeader) == 0 {
		return nil, fmt.Errorf("header row is empty")
	}

	cols := buildColumnMeta(rawHeader, schema)
	headerNames := make([]string, len(cols))
	for i := range cols {
		headerNames[i] = cols[i].Name
	}

	return &modelBuilder{
		header:          cols,
		hasData:         make([]bool, len(cols)),
		rows:            make([]Row, 0, 1024),
		timeColumnIndex: featuretimewindow.FindTimeColumnIndex(headerNames),
		rowTimes:        make([]time.Time, 0, 1024),
		rowHasTimes:     make([]bool, 0, 1024),
	}, nil
}

func buildColumnMeta(rawHeader []string, schema ColumnSchema) []ui.ColumnMeta {
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

	return cols
}

func (b *modelBuilder) addRow(source []string, originalIndex int, copyCols bool) {
	rowCols := source
	if copyCols {
		rowCols = append([]string(nil), source...)
	}
	for colIdx := range b.header {
		if colIdx < len(rowCols) && strings.TrimSpace(rowCols[colIdx]) != "" {
			b.hasData[colIdx] = true
		}
	}

	row := Row{
		Cols:          rowCols,
		OriginalIndex: originalIndex,
	}
	row.ID = row.ComputeID()
	b.rows = append(b.rows, row)

	ts, ok := b.parseRowTime(rowCols)
	b.rowTimes = append(b.rowTimes, ts)
	b.rowHasTimes = append(b.rowHasTimes, ok)
	if ok {
		if !b.hasTimeBounds {
			b.timeMin = ts
			b.timeMax = ts
			b.hasTimeBounds = true
		} else {
			if ts.Before(b.timeMin) {
				b.timeMin = ts
			}
			if ts.After(b.timeMax) {
				b.timeMax = ts
			}
		}
	}
}

func (b *modelBuilder) parseRowTime(cols []string) (time.Time, bool) {
	if b.timeColumnIndex < 0 || b.timeColumnIndex >= len(cols) {
		return time.Time{}, false
	}
	return featuretimewindow.ParseLogTimestamp(cols[b.timeColumnIndex])
}

func (b *modelBuilder) finish() *Model {
	for i := range b.header {
		if b.hasData[i] {
			continue
		}
		if b.header[i].Role != ui.RolePrimary {
			b.header[i].Visible = false
		}
		b.header[i].Weight = 0
		b.header[i].Width = 0
	}

	return &Model{
		table: tableState{
			header:          b.header,
			rows:            b.rows,
			markedRows:      make(map[uint64]ui.MarkColor),
			commentRows:     make(map[uint64]string),
			sortColumn:      -1,
			rowOrder:        makeIndexRange(len(b.rows)),
			searchColumns:   buildSearchColumnOrder(b.header),
			timeColumnIndex: b.timeColumnIndex,
			rowTimes:        b.rowTimes,
			rowHasTimes:     b.rowHasTimes,
			hasTimeBounds:   b.hasTimeBounds,
			timeMin:         b.timeMin,
			timeMax:         b.timeMax,
			derivedTimeData: true,
		},
		view: viewState{mode: modeView},
	}
}

func buildSearchColumnOrder(cols []ui.ColumnMeta) []int {
	if len(cols) == 0 {
		return nil
	}

	order := make([]int, 0, len(cols))
	seen := make(map[int]struct{}, len(cols))

	for _, wantRole := range []ui.ColumnRole{ui.RolePrimary, ui.RoleNormal, ui.RoleSecondary} {
		for _, col := range cols {
			if col.Role != wantRole {
				continue
			}
			if _, ok := seen[col.Index]; ok {
				continue
			}
			seen[col.Index] = struct{}{}
			order = append(order, col.Index)
		}
	}

	for _, col := range cols {
		if _, ok := seen[col.Index]; ok {
			continue
		}
		order = append(order, col.Index)
	}

	return order
}

func makeIndexRange(n int) []int {
	if n <= 0 {
		return nil
	}
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}
