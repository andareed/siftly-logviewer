package siftly

import (
	featuregraph "github.com/andareed/siftly-hostlog/internal/siftly/features/graph"
	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

type graphRenderCache struct {
	valid       bool
	dataVersion uint64
	rowsLen     int
	contentW    int
	graphW      int
	graphH      int
	timeCol     int
	seriesCol   int
	valueCol    int
	maxKeys     int
	scaleMode   string
	aggregate   string
	layout      string
	fillMode    string
	prepared    featuregraph.Prepared
}

type viewState struct {
	mode                    mode
	command                 CommandInput
	modeHintSeq             int
	prefixHintSeq           int
	graphWindow             featuregraph.Window
	graphDataVersion        uint64
	graphCache              graphRenderCache
	drawerOpen              bool
	drawerHeight            int
	notice                  ui.NoticeState
	searchQuery             string
	lastColumnsSpec         string
	pendingViewPrefix       string
	visibleStart            int
	visibleEnd              int
	debugCursorHeight       int
	debugHeightFree         int
	debugDesiredAboveHeight int
	rowHeights              map[int]int
	timeWindow              featuretimewindow.UIState
}
