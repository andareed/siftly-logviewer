package todaylog

import (
	"fmt"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/shared/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(statsFile, debugLogPath, filterPresetsPath, filterHistoryPath string) error {
	cleanup, err := logging.SetupLogging(debugLogPath)
	if err != nil {
		return fmt.Errorf("Setup logging: %w", err)
	}
	defer cleanup()

	logging.Info("siftly-todaylog: Started")
	m, err := LoadModelAuto(statsFile)

	if err != nil {
		return fmt.Errorf("Loading %q: %w", statsFile, err)
	}
	configureFilterConfig(m, filterPresetsPath, filterHistoryPath)

	if _, err = tea.NewProgram(m, tui.ProgramOptions()...).Run(); err != nil {
		return fmt.Errorf("TodayLog Program: %w", err)
	}
	return nil
}
