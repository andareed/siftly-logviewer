package siftly

import "github.com/andareed/siftly-hostlog/internal/siftly/ui"

// FullSourceReloadFunc reloads a model from its original source without an
// initial source-level reduction such as todaylog's prefilter.
type FullSourceReloadFunc func() (*Model, error)

func (m *Model) SetFullSourceReload(fn FullSourceReloadFunc) {
	m.fullSourceReload = fn
}

func (m *Model) CanReloadFullSource() bool {
	return m.fullSourceReload != nil
}

func mergeAnnotationsInto(dst *Model, src *Model) {
	if dst == nil || src == nil {
		return
	}

	present := make(map[uint64]struct{}, len(dst.table.rows))
	for _, row := range dst.table.rows {
		present[row.ID] = struct{}{}
	}

	if dst.table.markedRows == nil {
		dst.table.markedRows = make(map[uint64]ui.MarkColor)
	}
	for id, mark := range src.table.markedRows {
		if _, ok := present[id]; ok {
			dst.table.markedRows[id] = mark
		}
	}

	if dst.table.commentRows == nil {
		dst.table.commentRows = make(map[uint64]string)
	}
	for id, comment := range src.table.commentRows {
		if _, ok := present[id]; ok {
			dst.table.commentRows[id] = comment
		}
	}
}
