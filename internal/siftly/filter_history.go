package siftly

import (
	"os"
	"strings"
)

func (m *Model) recordFilterHistory(pattern string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil
	}

	historyPath := strings.TrimSpace(m.filterConfig.HistoryPath)
	if historyPath == "" {
		return nil
	}

	cfg, err := LoadFilterHistoryConfig(historyPath)
	if err != nil && !isMissingFileError(err) {
		return err
	}
	cfg.History = prependUnique(cfg.History, pattern)
	if len(cfg.History) > maxFilterHistory {
		cfg.History = cfg.History[:maxFilterHistory]
	}
	return SaveFilterHistoryConfig(historyPath, cfg)
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

func isMissingFileError(err error) bool {
	return os.IsNotExist(err)
}
