package siftly

import (
	"encoding/json"
	"os"
	"strings"
)

const (
	filterConfigFile = "filters.json"
	maxFilterHistory = 50
)

type FilterConfig struct {
	Presets PresetList `json:"presets"`
	History []string   `json:"history"`
}

type Preset struct {
	Pattern     string `json:"pattern"`
	Description string `json:"description,omitempty"`
}

type PresetList []Preset

func filterConfigPath() string {
	return filterConfigFile
}

func defaultFilterConfig() FilterConfig {
	return FilterConfig{
		Presets: PresetList{
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
		},
		History: []string{},
	}
}

func LoadFilterConfig(path string) (FilterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultFilterConfig(), nil
		}
		return FilterConfig{}, err
	}

	var cfg FilterConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FilterConfig{}, err
	}
	normalizeFilterConfig(&cfg)
	return cfg, nil
}

func SaveFilterConfig(path string, cfg FilterConfig) error {
	normalizeFilterConfig(&cfg)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func normalizeFilterConfig(cfg *FilterConfig) {
	if cfg == nil {
		return
	}
	cfg.Presets = normalizePresetList(cfg.Presets)
	cfg.History = normalizeFilterList(cfg.History)
}

func normalizePresetList(list PresetList) PresetList {
	seen := make(map[string]struct{}, len(list))
	out := make(PresetList, 0, len(list))
	for _, item := range list {
		pattern := strings.TrimSpace(item.Pattern)
		if pattern == "" {
			continue
		}
		if _, ok := seen[pattern]; ok {
			continue
		}
		seen[pattern] = struct{}{}
		item.Pattern = pattern
		item.Description = strings.TrimSpace(item.Description)
		out = append(out, item)
	}
	return out
}

func normalizeFilterList(list []string) []string {
	seen := make(map[string]struct{}, len(list))
	out := make([]string, 0, len(list))
	for _, item := range list {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func (p *PresetList) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	items := make([]Preset, 0, len(raw))
	for _, entry := range raw {
		var s string
		if err := json.Unmarshal(entry, &s); err == nil {
			items = append(items, Preset{Pattern: s})
			continue
		}
		var obj Preset
		if err := json.Unmarshal(entry, &obj); err != nil {
			return err
		}
		items = append(items, obj)
	}
	*p = items
	return nil
}

func (p PresetList) MarshalJSON() ([]byte, error) {
	items := make([]Preset, 0, len(p))
	items = append(items, p...)
	return json.Marshal(items)
}
