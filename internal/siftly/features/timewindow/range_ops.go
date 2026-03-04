package timewindow

import "time"

func ShiftRange(start, end, min, max time.Time, delta time.Duration) (time.Time, time.Time) {
	rangeDur := max.Sub(min)
	windowDur := end.Sub(start)
	if windowDur <= 0 {
		windowDur = StepMin
	}
	if windowDur > rangeDur {
		return min, max
	}

	nextStart := start.Add(delta)
	nextEnd := end.Add(delta)
	if nextStart.Before(min) {
		nextStart = min
		nextEnd = min.Add(windowDur)
	}
	if nextEnd.After(max) {
		nextEnd = max
		nextStart = max.Add(-windowDur)
	}
	return nextStart, nextEnd
}

func ExpandRange(start, end, min, max time.Time, delta time.Duration) (time.Time, time.Time) {
	if delta < 0 {
		nextStart := start.Add(delta)
		if nextStart.Before(min) {
			nextStart = min
		}
		start = nextStart
		if start.After(end) {
			end = start
		}
		return start, end
	}
	if delta > 0 {
		nextEnd := end.Add(delta)
		if nextEnd.After(max) {
			nextEnd = max
		}
		end = nextEnd
		if end.Before(start) {
			start = end
		}
	}
	return start, end
}

func AdjustStep(step time.Duration, increase bool) time.Duration {
	step = NormalizeStep(step)
	if increase {
		step *= 2
	} else {
		step /= 2
	}
	return NormalizeStep(step)
}
