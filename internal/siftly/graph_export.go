package siftly

import (
	"fmt"
	"os"

	featuregraph "github.com/andareed/siftly-hostlog/internal/siftly/features/graph"
)

func ExportGraphModel(m *Model, path string) error {
	if m == nil {
		return fmt.Errorf("model is nil")
	}
	if !m.graphConfig.Enabled {
		return fmt.Errorf("graph is not configured")
	}
	m.ensureTableDerivedState()

	header := make([]string, len(m.table.header))
	for i := range m.table.header {
		header[i] = m.table.header[i].Name
	}
	timeCol, seriesCol, valueCol, ok := featuregraph.ResolveColumnIndices(header, m.graphConfig)
	if !ok {
		return fmt.Errorf("graph columns not configured")
	}

	rows := make([][]string, len(m.table.rows))
	for i := range m.table.rows {
		rows[i] = m.table.rows[i].Cols
	}
	data, err := featuregraph.ExportSVG(featuregraph.Input{
		Rows:            rows,
		FilteredIndices: m.table.filteredIndices,
		TimeColumn:      timeCol,
		SeriesColumn:    seriesCol,
		ValueColumn:     valueCol,
		MaxKeys:         m.view.graphWindow.MaxKeysOrDefault(),
		ScaleMode:       m.graphConfig.ScaleMode,
		AggregateMode:   m.graphConfig.Aggregate,
		LayoutMode:      m.graphConfig.Layout,
		FillMode:        m.graphConfig.FillMode,
	}, featuregraph.ExportOptions{
		Title: defaultGraphExportTitle(*m),
	})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
