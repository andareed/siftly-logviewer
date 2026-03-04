package devfmt

import (
	"fmt"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(inputPath, debugLogPath string, q Query) error {
	cleanup, err := logging.SetupLogging(debugLogPath)
	if err != nil {
		return fmt.Errorf("setup logging: %w", err)
	}
	defer cleanup()

	logging.Info("siftly-devfmt: Started")

	ds, mappings, err := LoadDataset(inputPath)
	if err != nil {
		return fmt.Errorf("load devinfo dump: %w", err)
	}

	rows := BuildRows(ds, mappings, q)
	m, err := BuildModel(inputPath, rows)
	if err != nil {
		return fmt.Errorf("build model: %w", err)
	}

	if _, err = tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("tea program: %w", err)
	}

	return nil
}
