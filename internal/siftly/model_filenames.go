package siftly

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const graphExportTimestampLayout = "2006-01-02_15-04-05"

func defaultExportName(m Model) string {
	if m.lastExportFileName != "" {
		return filepath.Base(m.lastExportFileName)
	}

	initialBase := filepath.Base(m.InitialPath)
	base := strings.TrimSuffix(initialBase, filepath.Ext(initialBase))
	if base == "" || base == "." || base == string(filepath.Separator) {
		base = "output"
	}
	return "export-" + base + ".csv"
}

func defaultGraphExportName(m Model) string {
	return defaultGraphExportNameAt(m, time.Now())
}

func defaultGraphExportNameAt(m Model, now time.Time) string {
	initialBase := filepath.Base(m.InitialPath)
	base := strings.TrimSuffix(initialBase, filepath.Ext(initialBase))
	if base == "" || base == "." || base == string(filepath.Separator) {
		base = "output"
	}
	return base + "-graph-" + now.Format(graphExportTimestampLayout) + ".svg"
}

func defaultGraphExportPath(m Model) string {
	return defaultGraphExportPathAt(m, time.Now())
}

func defaultGraphExportPathAt(m Model, now time.Time) string {
	name := defaultGraphExportNameAt(m, now)
	dir := defaultDialogDir(m)
	if dir == "" || dir == "." {
		return name
	}
	return filepath.Join(dir, name)
}

func defaultGraphExportTitle(m Model) string {
	initialBase := filepath.Base(m.InitialPath)
	base := strings.TrimSuffix(initialBase, filepath.Ext(initialBase))
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "Siftly graph export"
	}
	return base + " graph"
}

func defaultSaveName(m Model) string {
	// Case 1: we already have a current filename — use it.
	if m.fileName != "" {
		return filepath.Base(m.fileName)
	}

	// Case 2: otherwise fall back to whatever initial path we have.
	initial := m.InitialPath
	if initial == "" {
		return "output.json" // final fallback
	}

	// Case 3: if the initial path already ends with .json, use it.
	if strings.HasSuffix(strings.ToLower(initial), ".json") {
		return filepath.Base(initial)
	}

	// Case 4: replace any existing extension with .json.
	base := strings.TrimSuffix(filepath.Base(initial), filepath.Ext(initial))
	return base + ".json"
}

func defaultDialogDir(m Model) string {
	if strings.TrimSpace(m.InitialPath) != "" {
		path := m.InitialPath
		if !filepath.IsAbs(path) {
			if abs, err := filepath.Abs(path); err == nil {
				path = abs
			}
		}
		return filepath.Dir(path)
	}

	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
