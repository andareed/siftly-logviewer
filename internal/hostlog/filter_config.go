package hostlog

import (
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly"
)

const (
	defaultFilterPresetsPath = "hostlog-filters.json"
	defaultFilterHistoryPath = "hostlog-filter-history.json"
)

func configureFilterConfig(m *siftly.Model, presetsPath, historyPath string) {
	if strings.TrimSpace(presetsPath) == "" {
		presetsPath = defaultFilterPresetsPath
	}
	if strings.TrimSpace(historyPath) == "" {
		historyPath = defaultFilterHistoryPath
	}

	m.SetFilterConfig(siftly.FilterConfigSettings{
		DefaultPresets: hostlogDefaultFilterPresets(),
		PresetsPath:    presetsPath,
		HistoryPath:    historyPath,
	})
}

func hostlogDefaultFilterPresets() siftly.PresetList {
	return siftly.PresetList{
		{
			Pattern:     "(?i)OS Classification Score - -1|Function Classification Score - -1",
			Description: "Classification failures: score is -1 in Details",
		},
		{
			Pattern:     "(?i)Failed to learn .* : No updated classification\\.",
			Description: "Failed-to-learn details where classification was not updated",
		},
		{
			Pattern:     "(?i)Label Active Test.*DHTestLabel.*Executing action - Add Label\\. Details:",
			Description: "Long policy details: executing Add Label action",
		},
		{
			Pattern:     "(?i)Label Active Test.*Host evaluation changed from .*Duration: 5 minutes",
			Description: "Long evaluation details with transition and duration",
		},
		{
			Pattern:     "(?i)Assigned Label - Assigned Label no longer includes DHTestLabel; Context: Removed by plugin Advanced Tools",
			Description: "Property details for label removal context",
		},
		{
			Pattern:     "(?i)NIC Vendor Value - Property value cleared: NIC Vendor Value - .*; Context: Purger",
			Description: "Property value cleared events from Purger context",
		},
	}
}
