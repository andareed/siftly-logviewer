package hostlog

import (
	"fmt"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(inputPath, debugLogPath, filterPresetsPath, filterHistoryPath string) error {
	cleanup, err := logging.SetupLogging(debugLogPath)
	if err != nil {
		return fmt.Errorf("setup logging: %w", err)
	}
	defer cleanup()

	logging.Info("siftly-hostlog: Started")

	m, err := LoadModelAuto(inputPath)
	if err != nil {
		return fmt.Errorf("load %q: %w", inputPath, err)
	}
	configureFilterConfig(m, filterPresetsPath, filterHistoryPath)

	if _, err = tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("tea program: %w", err)
	}

	return nil
}
