package siftly

import (
	"regexp"
	"strings"
)

// Row contains only persisted/source data. It has no rendering concerns.
type Row struct {
	Cols          []string
	ID            uint64
	OriginalIndex int // Row number in source data (1-based after header).
}

func (r Row) ComputeID() uint64 {
	var h uint64 = 14695981039346656037
	for _, col := range r.Cols {
		norm := strings.ToLower(strings.TrimSpace(col))
		for i := 0; i < len(norm); i++ {
			h ^= uint64(norm[i])
			h *= 1099511628211
		}
		h *= 1099511628211
	}
	return h
}

func (r Row) Join(sep string) string {
	var b strings.Builder
	for i, col := range r.Cols {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(col)
	}
	return b.String()
}

func (r Row) MatchesColumns(re *regexp.Regexp, order []int) bool {
	if re == nil {
		return true
	}

	if len(order) == 0 {
		for _, col := range r.Cols {
			if re.MatchString(col) {
				return true
			}
		}
		return false
	}

	for _, idx := range order {
		if idx < 0 || idx >= len(r.Cols) {
			continue
		}
		if re.MatchString(r.Cols[idx]) {
			return true
		}
	}
	return false
}

// String implements fmt.Stringer using tab as the default delimiter.
func (r Row) String() string {
	return r.Join("\t")
}
