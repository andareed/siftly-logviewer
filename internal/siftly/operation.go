package siftly

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const operationTickInterval = time.Second

type operationState struct {
	active  bool
	seq     int
	label   string
	started time.Time
}

type operationTickMsg struct {
	ID int
}

type saveCompleteMsg struct {
	ID       int
	Path     string
	Duration time.Duration
	Err      error
}

type graphExportCompleteMsg struct {
	ID       int
	Path     string
	Duration time.Duration
	Err      error
}

type filterCompleteMsg struct {
	ID              int
	FilteredIndices []int
	CurrentRowHash  uint64
	DoneMessage     string
	DoneKind        string
	Duration        time.Duration
}

type fullSourceReloadCompleteMsg struct {
	ID             int
	Model          *Model
	CurrentRowHash uint64
	Duration       time.Duration
	Err            error
}

func (m *Model) beginOperation(label string) (int, tea.Cmd) {
	m.view.operation.seq++
	m.view.operation.active = true
	m.view.operation.label = label
	m.view.operation.started = time.Now()
	m.view.notice.Set(operationNotice(label, 0), "info")
	return m.view.operation.seq, operationTickCmd(m.view.operation.seq)
}

func operationTickCmd(id int) tea.Cmd {
	return tea.Tick(operationTickInterval, func(time.Time) tea.Msg {
		return operationTickMsg{ID: id}
	})
}

func (m *Model) handleOperationTick(msg operationTickMsg) tea.Cmd {
	if !m.view.operation.active || msg.ID != m.view.operation.seq {
		return nil
	}
	elapsed := time.Since(m.view.operation.started)
	m.view.notice.Set(operationNotice(m.view.operation.label, elapsed), "info")
	return operationTickCmd(msg.ID)
}

func (m *Model) finishOperation(id int) bool {
	if !m.view.operation.active || id != m.view.operation.seq {
		return false
	}
	m.view.operation.active = false
	m.view.operation.label = ""
	m.view.operation.started = time.Time{}
	return true
}

func operationNotice(label string, elapsed time.Duration) string {
	if elapsed < operationTickInterval {
		return label + "..."
	}
	return fmt.Sprintf("%s... %s", label, elapsed.Round(time.Second))
}

func (m *Model) startSaveOperation(path string) tea.Cmd {
	id, tick := m.beginOperation("Saving")
	saveCmd := func() tea.Msg {
		start := time.Now()
		err := SaveModel(m, path)
		return saveCompleteMsg{
			ID:       id,
			Path:     path,
			Duration: time.Since(start),
			Err:      err,
		}
	}
	return tea.Batch(tick, saveCmd)
}

func (m *Model) handleSaveComplete(msg saveCompleteMsg) tea.Cmd {
	if !m.finishOperation(msg.ID) {
		return nil
	}
	if msg.Err != nil {
		return m.view.notice.Start("Save error", "error", noticeDuration)
	}
	m.fileName = msg.Path
	m.dirty = false
	return m.view.notice.Start(fmt.Sprintf("Saved in %s", msg.Duration.Round(time.Millisecond)), "success", noticeDuration)
}

func (m *Model) startGraphExportOperation(path string) tea.Cmd {
	id, tick := m.beginOperation("Exporting graph")
	exportCmd := func() tea.Msg {
		start := time.Now()
		err := ExportGraphModel(m, path)
		return graphExportCompleteMsg{
			ID:       id,
			Path:     path,
			Duration: time.Since(start),
			Err:      err,
		}
	}
	return tea.Batch(tick, exportCmd)
}

func (m *Model) handleGraphExportComplete(msg graphExportCompleteMsg) tea.Cmd {
	if !m.finishOperation(msg.ID) {
		return nil
	}
	if msg.Err != nil {
		return m.view.notice.Start("Graph export error", "error", noticeDuration)
	}
	m.lastGraphExportFileName = msg.Path
	return m.view.notice.Start(fmt.Sprintf("Graph exported in %s", msg.Duration.Round(time.Millisecond)), "success", noticeDuration)
}

func (m *Model) startFilterOperation(doneMessage string) tea.Cmd {
	return m.startFilterOperationWithKind(doneMessage, "success")
}

func (m *Model) startFilterOperationWithKind(doneMessage, doneKind string) tea.Cmd {
	m.clearRowRangeSelection()
	m.ensureTableDerivedState()
	currentRowHash := m.currentRowHashID()
	job := filterJob{
		rows:            m.table.rows,
		rowOrder:        m.table.rowOrder,
		markedRows:      m.table.markedRows,
		showOnlyMarked:  m.table.showOnlyMarked,
		filterEnabled:   m.table.filterEnabled,
		filterRegex:     m.table.filterRegex,
		filterWholeRow:  m.table.filterWholeRow,
		searchColumns:   m.table.searchColumns,
		timeWindow:      m.table.timeWindow,
		rowTimes:        m.table.rowTimes,
		rowHasTimes:     m.table.rowHasTimes,
		timeWindowOn:    m.table.timeWindow.Enabled,
		currentRowHash:  currentRowHash,
		completionLabel: doneMessage,
	}
	if doneKind == "" {
		doneKind = "success"
	}
	id, tick := m.beginOperation("Filtering")
	filterCmd := func() tea.Msg {
		start := time.Now()
		return filterCompleteMsg{
			ID:              id,
			FilteredIndices: job.run(),
			CurrentRowHash:  job.currentRowHash,
			DoneMessage:     job.completionLabel,
			DoneKind:        doneKind,
			Duration:        time.Since(start),
		}
	}
	return tea.Batch(tick, filterCmd)
}

func (m *Model) handleFilterComplete(msg filterCompleteMsg) tea.Cmd {
	if !m.finishOperation(msg.ID) {
		return nil
	}

	m.invalidateSearchIndex()
	m.table.filteredIndices = msg.FilteredIndices
	if len(m.table.filteredIndices) == 0 {
		m.cursor = -1
	}
	m.jumpToHashID(msg.CurrentRowHash)
	m.clampCursor()
	m.bumpGraphDataVersion()
	m.refreshView("filter-complete", false)

	label := msg.DoneMessage
	if label == "" {
		label = "Filter applied"
	}
	notice := m.view.notice.Start(
		fmt.Sprintf("%s (%d rows, %s)", label, len(m.table.filteredIndices), msg.Duration.Round(time.Millisecond)),
		msg.DoneKind,
		noticeDuration,
	)
	return batchCmd(notice, m.ensureSearchIndex())
}

func (m *Model) startFullSourceReloadOperation() tea.Cmd {
	if m.fullSourceReload == nil {
		return m.view.notice.Start("No prefiltered source to reload", "warn", noticeDuration)
	}
	m.clearRowRangeSelection()

	reload := m.fullSourceReload
	currentRowHash := m.currentRowHashID()
	id, tick := m.beginOperation("Loading full dataset")
	reloadCmd := func() tea.Msg {
		start := time.Now()
		next, err := reload()
		return fullSourceReloadCompleteMsg{
			ID:             id,
			Model:          next,
			CurrentRowHash: currentRowHash,
			Duration:       time.Since(start),
			Err:            err,
		}
	}
	return tea.Batch(tick, reloadCmd)
}

func (m *Model) handleFullSourceReloadComplete(msg fullSourceReloadCompleteMsg) tea.Cmd {
	if !m.finishOperation(msg.ID) {
		return nil
	}
	if msg.Err != nil {
		return m.view.notice.Start("Reload error", "error", noticeDuration)
	}
	if msg.Model == nil {
		return m.view.notice.Start("Reload error", "error", noticeDuration)
	}

	previous := *m
	next := msg.Model
	mergeAnnotationsInto(next, &previous)
	next.filterConfig = previous.filterConfig
	next.fileName = previous.fileName
	next.lastExportFileName = previous.lastExportFileName
	next.lastGraphExportFileName = previous.lastGraphExportFileName
	next.dirty = previous.dirty
	next.terminalHeight = previous.terminalHeight
	next.terminalWidth = previous.terminalWidth
	next.ready = previous.ready
	if next.graphConfig.Enabled {
		next.view.graphWindow.Open = previous.view.graphWindow.Open
	}

	*m = *next
	m.applyFilter()
	m.jumpToHashID(msg.CurrentRowHash)
	m.clampCursor()
	m.refreshView("full-source-reload", true)

	return m.view.notice.Start(
		fmt.Sprintf("Loaded full dataset (%d rows, %s)", len(m.table.rows), msg.Duration.Round(time.Millisecond)),
		"success",
		noticeDuration,
	)
}
