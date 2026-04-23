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
	filteredIndices []int // to store the list of indicides that match the current regex
	sortEnabled     bool
	sortColumn      int
	sortDesc        bool
	timeWindow      featuretimewindow.Window
	timeMin         time.Time
	timeMax         time.Time
	hasTimeBounds   bool
	timeColumnIndex int
	rowTimes        []time.Time
	rowHasTimes     []bool
}
