package pluginlog

import (
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly"
)

const (
	defaultFilterPresetsPath = "pluginlog-filters.json"
	defaultFilterHistoryPath = "pluginlog-filter-history.json"
)

func configureFilterConfig(m *siftly.Model, presetsPath, historyPath string) {
	if strings.TrimSpace(presetsPath) == "" {
		presetsPath = defaultFilterPresetsPath
	}
	if strings.TrimSpace(historyPath) == "" {
		historyPath = defaultFilterHistoryPath
	}

	m.SetFilterConfig(siftly.FilterConfigSettings{
		DefaultPresets: pluginlogDefaultFilterPresets(),
		PresetsPath:    presetsPath,
		HistoryPath:    historyPath,
	})
}

func pluginlogDefaultFilterPresets() siftly.PresetList {
	return siftly.PresetList{}
}
