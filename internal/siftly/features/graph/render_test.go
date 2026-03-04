package graph

import (
	"math"
	"strings"
	"testing"
)

func TestNormalizeModeParsers(t *testing.T) {
	if got := NormalizeScaleMode(" LOG1P "); got != ScaleLog1P {
		t.Fatalf("NormalizeScaleMode=%q want=%q", got, ScaleLog1P)
	}
	if got := NormalizeScaleMode("SYmLoG"); got != ScaleSymLog {
		t.Fatalf("NormalizeScaleMode=%q want=%q", got, ScaleSymLog)
	}
	if got := NormalizeScaleMode("bad"); got != ScaleLinear {
		t.Fatalf("NormalizeScaleMode=%q want=%q", got, ScaleLinear)
	}

	if got := NormalizeAggregateMode(" avg "); got != AggregateAvg {
		t.Fatalf("NormalizeAggregateMode=%q want=%q", got, AggregateAvg)
	}
	if got := NormalizeAggregateMode("MAX"); got != AggregateMax {
		t.Fatalf("NormalizeAggregateMode=%q want=%q", got, AggregateMax)
	}
	if got := NormalizeAggregateMode("bad"); got != AggregateLast {
		t.Fatalf("NormalizeAggregateMode=%q want=%q", got, AggregateLast)
	}

	if got := NormalizeLayoutMode(" split "); got != LayoutSplit {
		t.Fatalf("NormalizeLayoutMode=%q want=%q", got, LayoutSplit)
	}
	if got := NormalizeLayoutMode("bad"); got != LayoutOverlay {
		t.Fatalf("NormalizeLayoutMode=%q want=%q", got, LayoutOverlay)
	}

	if got := NormalizeFillMode(" none "); got != FillNone {
		t.Fatalf("NormalizeFillMode=%q want=%q", got, FillNone)
	}
	if got := NormalizeFillMode("bad"); got != FillHold {
		t.Fatalf("NormalizeFillMode=%q want=%q", got, FillHold)
	}
}

func TestSampleSeriesByTimeAggregateModes(t *testing.T) {
	points := []point{
		{ts: 1, val: 10},
		{ts: 2, val: 20},
		{ts: 2, val: 40},
		{ts: 3, val: 30},
	}

	assertSeries := func(name string, mode AggregateMode, want []float64) {
		t.Helper()
		got := sampleSeriesByTime(points, 1, 3, 3, mode, FillNone)
		if len(got) != len(want) {
			t.Fatalf("%s len=%d want=%d", name, len(got), len(want))
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("%s[%d]=%v want=%v", name, i, got[i], want[i])
			}
		}
	}

	assertSeries("last", AggregateLast, []float64{10, 40, 30})
	assertSeries("avg", AggregateAvg, []float64{10, 30, 30})
	assertSeries("max", AggregateMax, []float64{10, 40, 30})
	assertSeries("min", AggregateMin, []float64{10, 20, 30})
}

func TestNormalizeSeriesLogHighlightsLowRange(t *testing.T) {
	series := [][]float64{{1, 9, 99}}
	linear, ok := normalizeSeries(series, ScaleLinear)
	if !ok {
		t.Fatalf("linear normalizeSeries returned no values")
	}
	logv, ok := normalizeSeries(series, ScaleLog1P)
	if !ok {
		t.Fatalf("log normalizeSeries returned no values")
	}

	midLinear := linear[0][1]
	midLog := logv[0][1]
	if !(midLog > midLinear) {
		t.Fatalf("expected log midpoint > linear midpoint, got log=%f linear=%f", midLog, midLinear)
	}
}

func TestPrepareSplitLayoutProducesExpectedHeight(t *testing.T) {
	rows := [][]string{
		{"1", "a", "10"},
		{"2", "a", "20"},
		{"1", "b", "100"},
		{"2", "b", "120"},
	}
	in := Input{
		Width:           12,
		Height:          7,
		Rows:            rows,
		FilteredIndices: []int{0, 1, 2, 3},
		Cursor:          1,
		TimeColumn:      0,
		SeriesColumn:    1,
		ValueColumn:     2,
		MaxKeys:         5,
		ScaleMode:       string(ScaleLinear),
		AggregateMode:   string(AggregateLast),
		LayoutMode:      string(LayoutSplit),
	}

	prepared := Prepare(in)
	if !prepared.withCursor {
		t.Fatalf("expected prepared graph to include cursor line")
	}
	rendered := prepared.Render(cursorTimestamp(in))
	lines := strings.Split(rendered, "\n")
	if len(lines) != in.Height {
		t.Fatalf("rendered line count=%d want=%d", len(lines), in.Height)
	}
}

func TestSampleSeriesByTimeUsesNaNBeforeFirstPoint(t *testing.T) {
	points := []point{{ts: 10, val: 3}}
	got := sampleSeriesByTime(points, 0, 10, 3, AggregateLast, FillHold)
	if !math.IsNaN(got[0]) || !math.IsNaN(got[1]) || got[2] != 3 {
		t.Fatalf("unexpected sampled values: %#v", got)
	}
}

func TestSampleSeriesByTimeFillModeNoneLeavesGaps(t *testing.T) {
	points := []point{
		{ts: 1, val: 10},
		{ts: 3, val: 30},
	}
	got := sampleSeriesByTime(points, 1, 5, 5, AggregateLast, FillNone)
	if got[0] != 10 {
		t.Fatalf("got[0]=%v want=10", got[0])
	}
	if !math.IsNaN(got[1]) || got[2] != 30 || !math.IsNaN(got[3]) || !math.IsNaN(got[4]) {
		t.Fatalf("expected gaps between/after sparse points, got=%#v", got)
	}
}

func TestSampleSeriesByTimeFillModeHoldCarriesForward(t *testing.T) {
	points := []point{
		{ts: 1, val: 10},
		{ts: 3, val: 30},
	}
	got := sampleSeriesByTime(points, 1, 5, 5, AggregateLast, FillHold)
	want := []float64{10, 10, 30, 30, 30}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%v want=%v", i, got[i], want[i])
		}
	}
}
