package siftly

import (
	"regexp"
	"strings"
	"unicode/utf8"
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
		h = hashNormalizedColumn(h, col)
	}
	return h
}

func hashNormalizedColumn(h uint64, col string) uint64 {
	if isASCIIString(col) {
		start, end := trimASCIISpaceBounds(col)
		for i := start; i < end; i++ {
			b := col[i]
			if 'A' <= b && b <= 'Z' {
				b += 'a' - 'A'
			}
			h = fnv1aByte(h, b)
		}
		return fnv1aByte(h, 0)
	}

	norm := strings.ToLower(strings.TrimSpace(col))
	for i := 0; i < len(norm); i++ {
		h = fnv1aByte(h, norm[i])
	}
	return fnv1aByte(h, 0)
}

func fnv1aByte(h uint64, b byte) uint64 {
	h ^= uint64(b)
	h *= 1099511628211
	return h
}

func isASCIIString(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func trimASCIISpaceBounds(s string) (start, end int) {
	start = 0
	end = len(s)
	for start < end && isASCIISpace(s[start]) {
		start++
	}
	for end > start && isASCIISpace(s[end-1]) {
		end--
	}
	return start, end
}

func isASCIISpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	default:
		return false
	}
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
