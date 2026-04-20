package devfmt

import (
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly"
)

const (
	defaultFilterPresetsPath = "devfmt-filters.json"
	defaultFilterHistoryPath = "devfmt-filter-history.json"
)

func configureFilterConfig(m *siftly.Model, presetsPath, historyPath string) {
	if strings.TrimSpace(presetsPath) == "" {
		presetsPath = defaultFilterPresetsPath
	}
	if strings.TrimSpace(historyPath) == "" {
		historyPath = defaultFilterHistoryPath
	}

	m.SetFilterConfig(siftly.FilterConfigSettings{
		DefaultPresets: devfmtDefaultFilterPresets(),
		PresetsPath:    presetsPath,
		HistoryPath:    historyPath,
	})
}

func devfmtDefaultFilterPresets() siftly.PresetList {
	return siftly.PresetList{}
}
