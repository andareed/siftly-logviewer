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
	TimeParser      func(cols []string, timeColumnIndex int) (time.Time, bool)
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
	timeParser      func(cols []string, timeColumnIndex int) (time.Time, bool)
	rowTimes        []time.Time
	rowHasTimes     []bool
	hasTimeBounds   bool
	timeMin         time.Time
	timeMax         time.Time
	profile         bool
	rowsCapGrowths  int
	rowsGrowTime    time.Duration
	rowsCopyElems   int
	timesCapGrowths int
	timesGrowTime   time.Duration
	timesCopyElems  int
	flagsCapGrowths int
	flagsGrowTime   time.Duration
	flagsCopyElems  int
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

func (b *ModelBuilder) ReserveRows(capacity int) {
	if b == nil || b.builder == nil {
		return
	}
	b.builder.reserveRows(capacity)
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
		timeParser:      schema.TimeParser,
		rowTimes:        make([]time.Time, 0, 1024),
		rowHasTimes:     make([]bool, 0, 1024),
		profile:         logging.IsDebugMode(),
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
		if colIdx < len(rowCols) && rowCols[colIdx] != "" {
			b.hasData[colIdx] = true
		}
	}

	row := Row{
		Cols:          rowCols,
		OriginalIndex: originalIndex,
	}
	row.ID = row.ComputeID()

	b.appendRow(row)

	ts, ok := b.parseRowTime(rowCols)
	b.appendTimeMeta(ts, ok)

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
	if b.timeParser != nil {
		return b.timeParser(cols, b.timeColumnIndex)
	}
	return featuretimewindow.ParseLogTimestamp(cols[b.timeColumnIndex])
}

func (b *modelBuilder) finish() *Model {
	if b.profile {
		logging.Infof(
			"model builder storage: rows=%d rowsCap=%d grows=%d growTime=%s copyElems=%d rowTimesCap=%d grows=%d growTime=%s copyElems=%d rowHasTimesCap=%d grows=%d growTime=%s copyElems=%d",
			len(b.rows),
			cap(b.rows),
			b.rowsCapGrowths,
			b.rowsGrowTime.Round(time.Millisecond),
			b.rowsCopyElems,
			cap(b.rowTimes),
			b.timesCapGrowths,
			b.timesGrowTime.Round(time.Millisecond),
			b.timesCopyElems,
			cap(b.rowHasTimes),
			b.flagsCapGrowths,
			b.flagsGrowTime.Round(time.Millisecond),
			b.flagsCopyElems,
		)
	}
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

func (b *modelBuilder) reserveRows(capacity int) {
	if capacity <= 0 {
		return
	}
	if cap(b.rows) < capacity {
		rows := make([]Row, len(b.rows), capacity)
		copy(rows, b.rows)
		b.rows = rows
	}
	if cap(b.rowTimes) < capacity {
		rowTimes := make([]time.Time, len(b.rowTimes), capacity)
		copy(rowTimes, b.rowTimes)
		b.rowTimes = rowTimes
	}
	if cap(b.rowHasTimes) < capacity {
		rowHasTimes := make([]bool, len(b.rowHasTimes), capacity)
		copy(rowHasTimes, b.rowHasTimes)
		b.rowHasTimes = rowHasTimes
	}
}

func (b *modelBuilder) appendRow(row Row) {
	if !b.profile || len(b.rows) < cap(b.rows) {
		b.rows = append(b.rows, row)
		return
	}

	prevLen := len(b.rows)
	start := time.Now()
	b.rows = append(b.rows, row)
	b.rowsGrowTime += time.Since(start)
	b.rowsCapGrowths++
	b.rowsCopyElems += prevLen
}

func (b *modelBuilder) appendTimeMeta(ts time.Time, ok bool) {
	if !b.profile || (len(b.rowTimes) < cap(b.rowTimes) && len(b.rowHasTimes) < cap(b.rowHasTimes)) {
		b.rowTimes = append(b.rowTimes, ts)
		b.rowHasTimes = append(b.rowHasTimes, ok)
		return
	}

	if len(b.rowTimes) == cap(b.rowTimes) {
		prevLen := len(b.rowTimes)
		start := time.Now()
		b.rowTimes = append(b.rowTimes, ts)
		b.timesGrowTime += time.Since(start)
		b.timesCapGrowths++
		b.timesCopyElems += prevLen
	} else {
		b.rowTimes = append(b.rowTimes, ts)
	}

	if len(b.rowHasTimes) == cap(b.rowHasTimes) {
		prevLen := len(b.rowHasTimes)
		start := time.Now()
		b.rowHasTimes = append(b.rowHasTimes, ok)
		b.flagsGrowTime += time.Since(start)
		b.flagsCapGrowths++
		b.flagsCopyElems += prevLen
	} else {
		b.rowHasTimes = append(b.rowHasTimes, ok)
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
