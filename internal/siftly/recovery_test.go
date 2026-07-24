package siftly

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andareed/siftly-hostlog/internal/siftly/features/dialogs"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func TestRecoverySnapshotPromptsAndRestoresOnlyAfterConfirmation(t *testing.T) {
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
	m.table.header[0].Frozen = true
	m.recordChange("freeze column")
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
	if reopened.changes.pendingRecovery == nil {
		t.Fatal("matching recovery should be held pending")
	}
	if reopened.dirty {
		t.Fatal("finding recovery must not make the newly opened model dirty")
	}
	if reopened.table.markedRows[rowID] != ui.MarkNone || reopened.table.commentRows[rowID] != "" {
		t.Fatal("recovery was applied before user confirmation")
	}
	if _, ok := reopened.activeDialog.(*dialogs.Recovery); !ok {
		t.Fatalf("startup dialog = %T, want recovery prompt", reopened.activeDialog)
	}

	cmd, handled := reopened.handleDialogInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if !handled || cmd == nil {
		t.Fatalf("restore key handled=%t cmd=%v", handled, cmd != nil)
	}
	if reopened.changes.pendingRecovery != nil || !reopened.dirty {
		t.Fatalf("recovery was not applied after confirmation: pending=%t dirty=%t", reopened.changes.pendingRecovery != nil, reopened.dirty)
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
	if !reopened.table.header[0].Frozen {
		t.Fatal("frozen column state was not recovered")
	}

	if err := os.WriteFile(sourcePath, []byte("first,second\nbackend-changed,999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stale := newRecoveryTestModel(sourcePath, recoveryPath)
	if stale.changes.pendingRecovery != nil || stale.dirty {
		t.Fatal("recovery should be rejected after the backend source changes")
	}
	if _, err := os.Stat(recoveryPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale recovery was not removed: %v", err)
	}
}

func TestRecoveryPathIsHiddenAndAdjacentToOpenedFile(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "today.log")
	if err := os.WriteFile(sourcePath, []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	m := &Model{InitialPath: sourcePath}

	path, err := m.recoveryFilePath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, ".today.log.siftly-recovery-"+recoveryOwnerToken()+".json")
	if path != want {
		t.Fatalf("recovery path = %q, want %q", path, want)
	}
}

func TestDiscardRecoveryKeepsModelCleanAndRemovesSidecar(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.csv")
	recoveryPath := filepath.Join(dir, ".source.csv.siftly-recovery-"+recoveryOwnerToken()+".json")
	if err := os.WriteFile(sourcePath, []byte("first,second\nunique-row-payload,2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	original := newRecoveryTestModel(sourcePath, recoveryPath)
	original.markCurrent(ui.MarkAmber)
	if err := original.writeRecoverySnapshotNow(); err != nil {
		t.Fatal(err)
	}

	reopened := newRecoveryTestModel(sourcePath, recoveryPath)
	if reopened.changes.pendingRecovery == nil {
		t.Fatal("matching recovery should be held pending")
	}
	cmd, handled := reopened.handleDialogInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !handled || cmd == nil {
		t.Fatalf("discard key handled=%t cmd=%v", handled, cmd != nil)
	}
	if reopened.dirty || reopened.changes.pendingRecovery != nil {
		t.Fatalf("discard left dirty=%t pending=%t", reopened.dirty, reopened.changes.pendingRecovery != nil)
	}
	if len(reopened.table.markedRows) != 0 {
		t.Fatalf("discard applied recovered marks: %#v", reopened.table.markedRows)
	}
	if _, err := os.Stat(recoveryPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("discard left recovery sidecar: %v", err)
	}
}

func TestLegacyCacheRecoveryMigratesToAdjacentSidecar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	sourcePath := filepath.Join(dir, "source.csv")
	if err := os.WriteFile(sourcePath, []byte("first,second\nunique-row-payload,2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	original := newRecoveryTestModel(sourcePath, "")
	original.markCurrent(ui.MarkGreen)
	snapshot, err := original.buildRecoverySnapshot()
	if err != nil {
		t.Fatal(err)
	}
	legacyPath, err := original.legacyRecoveryFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeRecoveryFile(legacyPath, snapshot); err != nil {
		t.Fatal(err)
	}

	reopened := newRecoveryTestModel(sourcePath, "")
	sidecarPath, err := reopened.recoveryFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if reopened.changes.pendingRecovery == nil {
		t.Fatal("legacy recovery was not offered to the user")
	}
	if reopened.changes.pendingRecovery.path != sidecarPath {
		t.Fatalf("pending recovery path = %q, want migrated sidecar %q", reopened.changes.pendingRecovery.path, sidecarPath)
	}
	if _, err := os.Stat(sidecarPath); err != nil {
		t.Fatalf("migrated sidecar: %v", err)
	}
	if _, err := os.Stat(legacyPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("legacy recovery was not removed after migration: %v", err)
	}
	info, err := os.Stat(sidecarPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("sidecar permissions = %o, want 600", info.Mode().Perm())
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
	if recoveryPath != "" {
		m.SetRecoveryPath(recoveryPath)
	}
	m.InitialiseView()
	m.applyFilter()
	return m
}
