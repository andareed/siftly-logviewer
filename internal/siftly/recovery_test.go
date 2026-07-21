package siftly

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

func TestRecoverySnapshotRestoresOnlyAgainstIdenticalSource(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.csv")
	recoveryPath := filepath.Join(dir, "recovery.json")
	if err := os.WriteFile(sourcePath, []byte("first,second\nunique-row-payload,2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	m := newRecoveryTestModel(sourcePath, recoveryPath)
	m.markCurrent(ui.MarkRed)
	m.addComment("needs investigation")
	_ = runColumnOrderCommand(m, "second, first")
	if err := m.setSortSpec("second desc"); err != nil {
		t.Fatal(err)
	}
	if err := m.setFilterPattern("unique"); err != nil {
		t.Fatal(err)
	}
	if err := m.writeRecoverySnapshotNow(); err != nil {
		t.Fatalf("write recovery: %v", err)
	}

	data, err := os.ReadFile(recoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "unique-row-payload") {
		t.Fatal("recovery snapshot must not duplicate source row data")
	}

	reopened := newRecoveryTestModel(sourcePath, recoveryPath)
	rowID := reopened.table.rows[0].ID
	if !reopened.changes.recoveredOnStart || !reopened.dirty {
		t.Fatalf("recovery was not applied: recovered=%t dirty=%t", reopened.changes.recoveredOnStart, reopened.dirty)
	}
	if reopened.table.markedRows[rowID] != ui.MarkRed || reopened.table.commentRows[rowID] != "needs investigation" {
		t.Fatalf("annotations were not recovered: mark=%q comment=%q", reopened.table.markedRows[rowID], reopened.table.commentRows[rowID])
	}
	if !reopened.table.sortEnabled || reopened.table.filterPattern != "unique" {
		t.Fatalf("view state was not recovered: sort=%t filter=%q", reopened.table.sortEnabled, reopened.table.filterPattern)
	}
	if got := reopened.table.header[0].Name; got != "second" {
		t.Fatalf("column order was not recovered: first column=%q", got)
	}

	if err := os.WriteFile(sourcePath, []byte("first,second\nbackend-changed,999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stale := newRecoveryTestModel(sourcePath, recoveryPath)
	if stale.changes.recoveredOnStart || stale.dirty {
		t.Fatal("recovery should be rejected after the backend source changes")
	}
	if _, err := os.Stat(recoveryPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale recovery was not removed: %v", err)
	}
}

func TestAtomicWritePreservesExistingFileOnFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("write failed")
	err := writeAtomicFile(path, 0o600, func(w io.Writer) error {
		_, _ = w.Write([]byte("partial"))
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("write error = %v want %v", err, wantErr)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("atomic failure replaced target with %q", data)
	}
}

func TestCleanQuitRemovesRecoveryAfterUndoToBaseline(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.csv")
	recoveryPath := filepath.Join(dir, "recovery.json")
	if err := os.WriteFile(sourcePath, []byte("first,second\nunique-row-payload,2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	m := newRecoveryTestModel(sourcePath, recoveryPath)
	m.markCurrent(ui.MarkGreen)
	if err := m.writeRecoverySnapshotNow(); err != nil {
		t.Fatal(err)
	}
	_ = m.undoLastChange()
	if m.dirty {
		t.Fatal("undo should return to clean baseline")
	}
	_ = m.confirmOrQuit()
	if _, err := os.Stat(recoveryPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("clean quit left stale recovery: %v", err)
	}
}

func newRecoveryTestModel(sourcePath, recoveryPath string) *Model {
	row := Row{Cols: []string{"unique-row-payload", "2"}, OriginalIndex: 1}
	row.ID = row.ComputeID()
	m := &Model{
		InitialPath: sourcePath,
		cursor:      0,
		table: tableState{
			header: []ui.ColumnMeta{
				{Name: "first", Index: 0, Visible: true, MinWidth: 8, Weight: 1},
				{Name: "second", Index: 1, Visible: true, MinWidth: 8, Weight: 1},
			},
			rows:            []Row{row},
			filteredIndices: []int{0},
			markedRows:      map[uint64]ui.MarkColor{},
			commentRows:     map[uint64]string{},
			sortColumn:      -1,
			rowOrder:        []int{0},
			searchColumns:   []int{0, 1},
		},
	}
	m.SetRecoveryPath(recoveryPath)
	m.InitialiseView()
	m.applyFilter()
	return m
}
