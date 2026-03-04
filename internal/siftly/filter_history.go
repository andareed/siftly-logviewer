package siftly

import "strings"

func (m *Model) recordFilterHistory(pattern string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil
	}
	cfg, err := LoadFilterConfig(filterConfigPath())
	if err != nil {
		return err
	}
	cfg.History = prependUnique(cfg.History, pattern)
	if len(cfg.History) > maxFilterHistory {
		cfg.History = cfg.History[:maxFilterHistory]
	}
	return SaveFilterConfig(filterConfigPath(), cfg)
}

func prependUnique(list []string, value string) []string {
	out := make([]string, 0, len(list)+1)
	out = append(out, value)
	for _, item := range list {
		if item == value {
			continue
		}
		out = append(out, item)
	}
	return out
}
