package todaylog

import (
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly"
)

const (
	defaultFilterPresetsPath = "todaylog-filters.json"
	defaultFilterHistoryPath = "todaylog-filter-history.json"
)

func configureFilterConfig(m *siftly.Model, presetsPath, historyPath string) {
	if strings.TrimSpace(presetsPath) == "" {
		presetsPath = defaultFilterPresetsPath
	}
	if strings.TrimSpace(historyPath) == "" {
		historyPath = defaultFilterHistoryPath
	}

	m.SetFilterConfig(siftly.FilterConfigSettings{
		DefaultPresets: todaylogDefaultFilterPresets(),
		PresetsPath:    presetsPath,
		HistoryPath:    historyPath,
	})
}

func todaylogDefaultFilterPresets() siftly.PresetList {
	return siftly.PresetList{}
}
