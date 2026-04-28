package siftly

import (
	"regexp"
	"time"

	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

type tableState struct {
	header          []ui.ColumnMeta // single row for column titles in headerview
	rows            []Row
	markedRows      map[uint64]ui.MarkColor // map row index to color code
	commentRows     map[uint64]string       // map row index to string to store comments
	showOnlyMarked  bool
	filterPattern   string
	filterEnabled   bool
	filterRegex     *regexp.Regexp
	filterWholeRow  bool
	filteredIndices []int // to store the list of indicides that match the current regex
	sortEnabled     bool
	sortColumn      int
	sortDesc        bool
	rowOrder        []int
	searchColumns   []int
	timeWindow      featuretimewindow.Window
	timeMin         time.Time
	timeMax         time.Time
	hasTimeBounds   bool
	timeColumnIndex int
	rowTimes        []time.Time
	rowHasTimes     []bool
	derivedTimeData bool
}

func (m *Model) ensureTableDerivedState() {
	if len(m.table.rowOrder) != len(m.table.rows) {
		m.table.rowOrder = makeIndexRange(len(m.table.rows))
	}

	if len(m.table.header) > 0 {
		if len(m.table.searchColumns) != len(m.table.header) {
			m.table.searchColumns = buildSearchColumnOrder(m.table.header)
		}
		return
	}

	if len(m.table.searchColumns) == 0 && len(m.table.rows) > 0 {
		m.table.searchColumns = makeIndexRange(len(m.table.rows[0].Cols))
	}
}
