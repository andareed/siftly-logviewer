package graph

import "strings"

// Config controls the optional graph pane in siftly.
// Graph is disabled unless Enabled is true.
type Config struct {
	Enabled      bool
	TimeColumn   string
	SeriesColumn string
	ValueColumn  string
	Height       int
	MaxKeys      int
	ScaleMode    string
	Aggregate    string
	Layout       string
	// FillMode controls what happens between sparse points:
	// "hold" carries the previous value forward, "none" leaves a gap.
	FillMode string
}

type ScaleMode string

const (
	ScaleLinear ScaleMode = "linear"
	ScaleLog1P  ScaleMode = "log1p"
	ScaleSymLog ScaleMode = "symlog"
)

type AggregateMode string

const (
	AggregateLast AggregateMode = "last"
	AggregateAvg  AggregateMode = "avg"
	AggregateMax  AggregateMode = "max"
	AggregateMin  AggregateMode = "min"
)

type LayoutMode string

const (
	LayoutOverlay LayoutMode = "overlay"
	LayoutSplit   LayoutMode = "split"
)

type FillMode string

const (
	FillHold FillMode = "hold"
	FillNone FillMode = "none"
)

func NormalizeScaleMode(raw string) ScaleMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(ScaleLog1P):
		return ScaleLog1P
	case string(ScaleSymLog):
		return ScaleSymLog
	default:
		return ScaleLinear
	}
}

func NormalizeAggregateMode(raw string) AggregateMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(AggregateAvg):
		return AggregateAvg
	case string(AggregateMax):
		return AggregateMax
	case string(AggregateMin):
		return AggregateMin
	default:
		return AggregateLast
	}
}

func NormalizeLayoutMode(raw string) LayoutMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(LayoutSplit):
		return LayoutSplit
	default:
		return LayoutOverlay
	}
}

func NormalizeFillMode(raw string) FillMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(FillNone):
		return FillNone
	default:
		return FillHold
	}
}
