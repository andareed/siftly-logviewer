package siftly

import (
	"reflect"
	"regexp"

	featuretimewindow "github.com/andareed/siftly-hostlog/internal/siftly/features/timewindow"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxUndoEntries = 50
	maxUndoBytes   = 16 << 20
)

type trackedColumn struct {
	Name     string
	Index    int
	Role     ui.ColumnRole
	Visible  bool
	Frozen   bool
	MinWidth int
	Weight   float64
}

type trackedState struct {
	Columns        []trackedColumn
	MarkedRows     map[uint64]ui.MarkColor
	CommentRows    map[uint64]string
	TimeWindow     featuretimewindow.Window
	SortEnabled    bool
	SortColumn     int
	SortDesc       bool
	FilterPattern  string
	FilterEnabled  bool
	FilterWholeRow bool
	ShowOnlyMarked bool
	GraphOpen      bool
}

type historyEntry struct {
	state trackedState
	label string
	size  int
}

type changeTracker struct {
	initialized          bool
	current              trackedState
	baseline             trackedState
	undo                 []historyEntry
	redo                 []historyEntry
	undoBytes            int
	redoBytes            int
	recoverySeq          int
	pendingRecoveryCmd   tea.Cmd
	recoveryPathOverride string
	recoveredOnStart     bool
}

func (m *Model) initializeChangeTracking() {
	override := m.changes.recoveryPathOverride
	state := m.captureTrackedState()
	m.changes = changeTracker{
		initialized:          true,
		current:              state,
		baseline:             state,
		recoveryPathOverride: override,
	}
	m.dirty = false
}

func (m *Model) captureTrackedState() trackedState {
	columns := make([]trackedColumn, len(m.table.header))
	for i, column := range m.table.header {
		columns[i] = trackedColumn{
			Name:     column.Name,
			Index:    column.Index,
			Role:     column.Role,
			Visible:  column.Visible,
			Frozen:   column.Frozen,
			MinWidth: column.MinWidth,
			Weight:   column.Weight,
		}
	}
	return trackedState{
		Columns:        columns,
		MarkedRows:     cloneMarkMap(m.table.markedRows),
		CommentRows:    cloneCommentMap(m.table.commentRows),
		TimeWindow:     m.table.timeWindow,
		SortEnabled:    m.table.sortEnabled,
		SortColumn:     m.table.sortColumn,
		SortDesc:       m.table.sortDesc,
		FilterPattern:  m.table.filterPattern,
		FilterEnabled:  m.table.filterEnabled,
		FilterWholeRow: m.table.filterWholeRow,
		ShowOnlyMarked: m.table.showOnlyMarked,
		GraphOpen:      m.view.graphWindow.Open,
	}
}

func (m *Model) recordChange(label string) {
	after := m.captureTrackedState()
	if !m.changes.initialized {
		m.changes.initialized = true
		m.changes.current = after
		m.changes.baseline = trackedState{}
		m.dirty = true
		m.scheduleRecovery()
		return
	}
	if trackedStatesEqual(m.changes.current, after) {
		return
	}

	m.pushUndo(historyEntry{
		state: m.changes.current,
		label: label,
		size:  trackedStateSize(m.changes.current),
	})
	m.changes.current = after
	m.changes.redo = nil
	m.changes.redoBytes = 0
	m.updateDirtyFromBaseline()
	m.scheduleRecovery()
}

func (m *Model) markSavedBaseline() {
	state := m.captureTrackedState()
	if !m.changes.initialized {
		m.changes.initialized = true
	}
	m.changes.current = state
	m.changes.baseline = state
	m.dirty = false
	m.changes.recoverySeq++
	m.changes.pendingRecoveryCmd = nil
}

func (m *Model) updateDirtyFromBaseline() {
	m.dirty = !persistentStatesEqual(m.changes.current, m.changes.baseline)
}

func (m *Model) canUndo() bool { return len(m.changes.undo) > 0 }
func (m *Model) canRedo() bool { return len(m.changes.redo) > 0 }

func (m *Model) undoLastChange() tea.Cmd {
	if !m.canUndo() {
		return m.view.notice.Start("Nothing to undo", "warn", noticeDuration)
	}

	entry := m.changes.undo[len(m.changes.undo)-1]
	m.changes.undo = m.changes.undo[:len(m.changes.undo)-1]
	m.changes.undoBytes -= entry.size
	current := m.captureTrackedState()
	m.pushRedo(historyEntry{state: current, label: entry.label, size: trackedStateSize(current)})
	m.restoreTrackedState(entry.state, true)
	m.changes.current = entry.state
	m.updateDirtyFromBaseline()
	m.scheduleRecovery()
	return m.view.notice.Start("Undid "+entry.label, "", noticeDuration)
}

func (m *Model) redoLastChange() tea.Cmd {
	if !m.canRedo() {
		return m.view.notice.Start("Nothing to redo", "warn", noticeDuration)
	}

	entry := m.changes.redo[len(m.changes.redo)-1]
	m.changes.redo = m.changes.redo[:len(m.changes.redo)-1]
	m.changes.redoBytes -= entry.size
	current := m.captureTrackedState()
	m.pushUndo(historyEntry{state: current, label: entry.label, size: trackedStateSize(current)})
	m.restoreTrackedState(entry.state, true)
	m.changes.current = entry.state
	m.updateDirtyFromBaseline()
	m.scheduleRecovery()
	return m.view.notice.Start("Redid "+entry.label, "", noticeDuration)
}

func (m *Model) restoreTrackedState(state trackedState, refresh bool) {
	currentRowID := m.currentRowHashID()
	m.view.columnScrollOffset = 0
	m.table.header = make([]ui.ColumnMeta, len(state.Columns))
	for i, column := range state.Columns {
		m.table.header[i] = ui.ColumnMeta{
			Name:     column.Name,
			Index:    column.Index,
			Role:     column.Role,
			Visible:  column.Visible,
			Frozen:   column.Frozen,
			MinWidth: column.MinWidth,
			Weight:   column.Weight,
		}
	}
	m.table.markedRows = cloneMarkMap(state.MarkedRows)
	m.table.commentRows = cloneCommentMap(state.CommentRows)
	m.table.timeWindow = state.TimeWindow
	m.table.sortEnabled = state.SortEnabled
	m.table.sortColumn = state.SortColumn
	m.table.sortDesc = state.SortDesc
	m.table.filterPattern = state.FilterPattern
	m.table.filterEnabled = state.FilterEnabled
	m.table.filterWholeRow = state.FilterWholeRow
	m.table.filterRegex = nil
	if state.FilterPattern != "" {
		m.table.filterRegex, _ = regexp.Compile(state.FilterPattern)
	}
	m.table.showOnlyMarked = state.ShowOnlyMarked
	m.table.searchColumns = buildSearchColumnOrder(m.table.header)
	if m.graphConfig.Enabled {
		m.view.graphWindow.Open = state.GraphOpen
	}

	m.rebuildRowOrder()
	if refresh || m.ready {
		m.applyFilter()
		m.jumpToHashID(currentRowID)
		m.clampCursor()
	}
	m.view.timeWindow.DraftStart = state.TimeWindow.Start
	m.view.timeWindow.DraftEnd = state.TimeWindow.End
	m.updateTimeWindowInputsFromDraft()
	if refresh && m.ready {
		m.refreshView("tracked-state-restore", true)
	}
}

func (m *Model) pushUndo(entry historyEntry) {
	m.changes.undo = append(m.changes.undo, entry)
	m.changes.undoBytes += entry.size
	for len(m.changes.undo) > 1 && (len(m.changes.undo) > maxUndoEntries || m.changes.undoBytes > maxUndoBytes) {
		m.changes.undoBytes -= m.changes.undo[0].size
		m.changes.undo = m.changes.undo[1:]
	}
}

func (m *Model) pushRedo(entry historyEntry) {
	m.changes.redo = append(m.changes.redo, entry)
	m.changes.redoBytes += entry.size
	for len(m.changes.redo) > 1 && (len(m.changes.redo) > maxUndoEntries || m.changes.redoBytes > maxUndoBytes) {
		m.changes.redoBytes -= m.changes.redo[0].size
		m.changes.redo = m.changes.redo[1:]
	}
}

func trackedStatesEqual(left, right trackedState) bool {
	return reflect.DeepEqual(left, right)
}

func persistentStatesEqual(left, right trackedState) bool {
	return reflect.DeepEqual(left.Columns, right.Columns) &&
		reflect.DeepEqual(left.MarkedRows, right.MarkedRows) &&
		reflect.DeepEqual(left.CommentRows, right.CommentRows) &&
		left.TimeWindow == right.TimeWindow
}

func trackedStateSize(state trackedState) int {
	size := len(state.Columns)*96 + len(state.MarkedRows)*32
	for _, comment := range state.CommentRows {
		size += 32 + len(comment)
	}
	size += len(state.FilterPattern) + 128
	return size
}

func cloneMarkMap(source map[uint64]ui.MarkColor) map[uint64]ui.MarkColor {
	result := make(map[uint64]ui.MarkColor, len(source))
	for id, mark := range source {
		result[id] = mark
	}
	return result
}

func cloneCommentMap(source map[uint64]string) map[uint64]string {
	result := make(map[uint64]string, len(source))
	for id, comment := range source {
		result[id] = comment
	}
	return result
}

func trackedColumnsEqualCurrent(columns []trackedColumn, header []ui.ColumnMeta) bool {
	if len(columns) != len(header) {
		return false
	}
	type columnKey struct {
		index int
		name  string
	}
	seen := make(map[columnKey]int, len(header))
	for _, column := range header {
		seen[columnKey{index: column.Index, name: column.Name}]++
	}
	for _, column := range columns {
		key := columnKey{index: column.Index, name: column.Name}
		if seen[key] == 0 {
			return false
		}
		seen[key]--
	}
	return true
}
