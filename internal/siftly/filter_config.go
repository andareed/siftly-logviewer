package siftly

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxFilterHistory = 50
)

type FilterConfig struct {
	Presets PresetList
	History []string
}

type FilterPresetConfig struct {
	Presets PresetList `json:"presets"`
}

type FilterHistoryConfig struct {
	History []string `json:"history"`
}

type FilterConfigSettings struct {
	DefaultPresets PresetList
	PresetsPath    string
	HistoryPath    string
}

type Preset struct {
	Pattern     string `json:"pattern"`
	Description string `json:"description,omitempty"`
}

type PresetList []Preset

func (m *Model) SetFilterConfig(settings FilterConfigSettings) {
	m.filterConfig = normalizeFilterConfigSettings(settings)
}

func normalizeFilterConfigSettings(settings FilterConfigSettings) FilterConfigSettings {
	settings.DefaultPresets = normalizePresetList(settings.DefaultPresets)
	settings.PresetsPath = strings.TrimSpace(settings.PresetsPath)
	settings.HistoryPath = strings.TrimSpace(settings.HistoryPath)
	return settings
}

func LoadFilterPresetConfig(path string) (FilterPresetConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FilterPresetConfig{}, err
	}

	var cfg FilterPresetConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FilterPresetConfig{}, err
	}
	cfg.Presets = normalizePresetList(cfg.Presets)
	return cfg, nil
}

func LoadFilterHistoryConfig(path string) (FilterHistoryConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FilterHistoryConfig{}, err
	}

	var cfg FilterHistoryConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FilterHistoryConfig{}, err
	}
	cfg.History = normalizeFilterList(cfg.History)
	return cfg, nil
}

func SaveFilterHistoryConfig(path string, cfg FilterHistoryConfig) error {
	cfg.History = normalizeFilterList(cfg.History)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}

func (m Model) loadFilterConfig() (FilterConfig, error) {
	settings := normalizeFilterConfigSettings(m.filterConfig)
	cfg := FilterConfig{
		Presets: append(PresetList(nil), settings.DefaultPresets...),
		History: []string{},
	}

	if settings.PresetsPath != "" {
		fileCfg, err := LoadFilterPresetConfig(settings.PresetsPath)
		if err != nil && !os.IsNotExist(err) {
			return FilterConfig{}, err
		}
		cfg.Presets = mergePresetLists(cfg.Presets, fileCfg.Presets)
	}

	if settings.HistoryPath != "" {
		historyCfg, err := LoadFilterHistoryConfig(settings.HistoryPath)
		if err != nil && !os.IsNotExist(err) {
			return FilterConfig{}, err
		}
		cfg.History = historyCfg.History
	}

	return cfg, nil
}

func mergePresetLists(base, overlay PresetList) PresetList {
	out := append(PresetList(nil), normalizePresetList(base)...)
	if len(overlay) == 0 {
		return out
	}

	indexByPattern := make(map[string]int, len(out))
	for i, item := range out {
		indexByPattern[item.Pattern] = i
	}

	for _, item := range normalizePresetList(overlay) {
		if idx, ok := indexByPattern[item.Pattern]; ok {
			out[idx] = item
			continue
		}
		indexByPattern[item.Pattern] = len(out)
		out = append(out, item)
	}

	return out
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
