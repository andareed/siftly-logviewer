package dialogs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const dialogFileListLimit = 12

type fileDialogState struct {
	TargetPath     string
	TargetDir      string
	FileLines      []string
	TopRightState  string
	StatusKind     string
	StatusMessage  string
	PrimaryAction  string
	PrimaryEnabled bool
}

func effectiveInputValue(input, placeholder string) string {
	val := strings.TrimSpace(input)
	if val == "" {
		val = strings.TrimSpace(placeholder)
	}
	return val
}

func resolveTargetPath(baseDir, input, placeholder string) (path string, hasValue bool) {
	val := effectiveInputValue(input, placeholder)
	if val == "" {
		return "", false
	}
	if baseDir == "" {
		baseDir = "."
	}
	if filepath.IsAbs(val) {
		return filepath.Clean(val), true
	}
	return filepath.Clean(filepath.Join(baseDir, val)), true
}

func readDirPreview(dir string, limit int) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{fmt.Sprintf("(cannot read directory: %v)", err)}
	}
	if len(entries) == 0 {
		return []string{"(empty folder)"}
	}

	sort.Slice(entries, func(i, j int) bool {
		left, right := entries[i], entries[j]
		if left.IsDir() != right.IsDir() {
			return left.IsDir()
		}
		return strings.ToLower(left.Name()) < strings.ToLower(right.Name())
	})

	if limit <= 0 {
		limit = len(entries)
	}

	count := len(entries)
	if count > limit {
		count = limit
	}
	lines := make([]string, 0, count+1)
	for i := 0; i < count; i++ {
		entry := entries[i]
		if entry.IsDir() {
			lines = append(lines, fmt.Sprintf("[D] %s/", entry.Name()))
			continue
		}
		lines = append(lines, fmt.Sprintf("[F] %s", entry.Name()))
	}
	if len(entries) > count {
		lines = append(lines, fmt.Sprintf("... %d more", len(entries)-count))
	}
	return lines
}

func resolveFileDialogState(baseDir, input, placeholder, primaryAction string) fileDialogState {
	st := fileDialogState{
		PrimaryAction:  primaryAction,
		PrimaryEnabled: false,
		TopRightState:  "INVALID",
		StatusKind:     "error",
		StatusMessage:  "✖ Invalid filename",
	}

	targetPath, ok := resolveTargetPath(baseDir, input, placeholder)
	if !ok {
		st.TargetDir = strings.TrimSpace(baseDir)
		if st.TargetDir == "" {
			st.TargetDir = "."
		}
		st.FileLines = []string{"(enter a file name to preview destination)"}
		return st
	}

	st.TargetPath = targetPath
	st.TargetDir = filepath.Dir(targetPath)
	st.FileLines = readDirPreview(st.TargetDir, dialogFileListLimit)

	info, err := os.Stat(targetPath)
	switch {
	case err == nil && info.IsDir():
		st.StatusKind = "error"
		st.StatusMessage = "✖ Invalid filename"
		st.TopRightState = "INVALID"
		st.PrimaryEnabled = false
		return st
	case err == nil:
		st.StatusKind = "warn"
		st.StatusMessage = "⚠ File exists - will overwrite"
		st.TopRightState = "OVERWRITE"
		st.PrimaryAction = "Overwrite"
		st.PrimaryEnabled = true
		return st
	case os.IsNotExist(err):
		st.StatusKind = "success"
		st.StatusMessage = "✓ New file"
		st.TopRightState = ""
		st.PrimaryAction = primaryAction
		st.PrimaryEnabled = true
		return st
	default:
		st.StatusKind = "error"
		st.StatusMessage = "✖ Invalid filename"
		st.TopRightState = "INVALID"
		st.PrimaryEnabled = false
		return st
	}
}
