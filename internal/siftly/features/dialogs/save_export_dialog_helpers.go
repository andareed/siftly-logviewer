package dialogs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
	"github.com/charmbracelet/lipgloss"
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

func preferredFileDialogWidth(title, input string, state fileDialogState) int {
	innerWidth := max(lipgloss.Width(title), lipgloss.Width(input))
	for _, value := range append([]string{state.TargetDir, state.StatusMessage}, state.FileLines...) {
		if width := lipgloss.Width(value); width > innerWidth {
			innerWidth = width
		}
	}
	return clampDialogWidth(innerWidth+4, 44, 78)
}

func fileDialogContent(inputView string, state fileDialogState, innerWidth, outerHeight int, tokens ui.DesignTokens) []string {
	contentLimit := outerHeight - 2
	if contentLimit < 4 {
		contentLimit = 4
	}
	action := renderDialogActionRowWithKeys(innerWidth, "Enter", state.PrimaryAction, state.PrimaryEnabled, "Esc", "Cancel", tokens)
	status := dialogStatusLine(state.StatusKind, state.StatusMessage, tokens)

	content := []string{dialogSectionLabel("Filename", tokens), inputView}
	tail := []string{"", status, action}
	remaining := contentLimit - len(content) - len(tail)
	if remaining >= 3 {
		content = append(content, "", dialogSectionLabel("Location", tokens), state.TargetDir)
		remaining -= 3
	}
	if remaining >= 3 {
		fileSlots := remaining - 2
		content = append(content, "", dialogSectionLabel("Files in folder", tokens))
		content = append(content, fitFileDialogLines(state.FileLines, fileSlots)...)
	}
	return append(content, tail...)
}

func fitFileDialogLines(lines []string, slots int) []string {
	if slots <= 0 || len(lines) == 0 {
		return nil
	}
	if len(lines) <= slots {
		return lines
	}
	if slots == 1 {
		return []string{fmt.Sprintf("... %d entries", len(lines))}
	}
	visible := append([]string(nil), lines[:slots-1]...)
	return append(visible, fmt.Sprintf("... %d more", len(lines)-slots+1))
}
