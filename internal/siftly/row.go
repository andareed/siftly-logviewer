package siftly

import (
	"hash/fnv"
	"strings"
)

// Row contains only persisted/source data. It has no rendering concerns.
type Row struct {
	Cols          []string
	ID            uint64
	OriginalIndex int // Row number in source data (1-based after header).
}

func (r Row) ComputeID() uint64 {
	h := fnv.New64a()
	for _, col := range r.Cols {
		norm := strings.ToLower(strings.TrimSpace(col))
		h.Write([]byte(norm))
		h.Write([]byte{0})
	}
	return h.Sum64()
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

// String implements fmt.Stringer using tab as the default delimiter.
func (r Row) String() string {
	return r.Join("\t")
}
