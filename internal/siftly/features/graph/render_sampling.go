package graph

import "math"

func sampleSeriesByTime(points []point, minTS, maxTS int64, width int, mode AggregateMode, fillMode FillMode) []float64 {
	if width <= 0 {
		return nil
	}
	out := make([]float64, width)
	for i := range out {
		out[i] = math.NaN()
	}
	if len(points) == 0 {
		return out
	}
	if maxTS <= minTS {
		maxTS = minTS + 1
	}

	idx := 0
	seen := false
	carry := 0.0
	for x := 0; x < width; x++ {
		target := minTS + int64(float64(maxTS-minTS)*float64(x)/float64(maxInt(1, width-1)))
		hasBucket := false
		bucket := 0.0
		count := 0
		for idx < len(points) && points[idx].ts <= target {
			v := points[idx].val
			if !hasBucket {
				bucket = v
				hasBucket = true
			} else {
				switch mode {
				case AggregateAvg:
					bucket += v
				case AggregateMax:
					if v > bucket {
						bucket = v
					}
				case AggregateMin:
					if v < bucket {
						bucket = v
					}
				default:
					bucket = v
				}
			}
			carry = v
			seen = true
			count++
			idx++
		}
		if hasBucket {
			if mode == AggregateAvg && count > 0 {
				bucket /= float64(count)
			}
			out[x] = bucket
			continue
		}
		// `hold` keeps a step-line between sparse updates; `none` leaves NaN gaps.
		if seen && fillMode == FillHold {
			out[x] = carry
		}
	}
	return out
}

func sampleIndex(length int, width int, x int) int {
	if length <= 1 || width <= 1 {
		return 0
	}
	idx := int(float64(x) * float64(length-1) / float64(width-1))
	if idx < 0 {
		return 0
	}
	if idx >= length {
		return length - 1
	}
	return idx
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
