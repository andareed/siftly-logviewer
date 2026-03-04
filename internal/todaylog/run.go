package todaylog

import (
	"fmt"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(statsFile, debugLogPath string) error {
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

	if _, err = tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("TodayLog Program: %w", err)
	}
	return nil
}
