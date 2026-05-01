package tui

import (
	"os"
	"strconv"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	tea "github.com/charmbracelet/bubbletea"
)

const defaultFPS = 24

func ProgramOptions() []tea.ProgramOption {
	fps := renderFPS()
	logging.Infof("bubbletea renderer fps=%d", fps)
	return []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithFPS(fps),
	}
}

func renderFPS() int {
	raw := strings.TrimSpace(os.Getenv("SIFTLY_FPS"))
	if raw == "" {
		return defaultFPS
	}
	fps, err := strconv.Atoi(raw)
	if err != nil || fps < 1 {
		logging.Warnf("invalid SIFTLY_FPS=%q, using default %d", raw, defaultFPS)
		return defaultFPS
	}
	return fps
}
