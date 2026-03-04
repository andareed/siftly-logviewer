package timewindow

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
)

const (
	InputLayout          = "2006-01-02 15:04:05"
	InputLayoutNoSeconds = "2006-01-02 15:04"
)

const (
	FocusStart = iota
	FocusEnd
	FocusScrubber
)

const (
	DrawerContentHeight = 5
	DrawerHeight        = DrawerContentHeight + 2
	StepMin             = 15 * time.Minute
	StepDefault         = 30 * time.Minute
	StepMax             = 2 * time.Hour
)

type Window struct {
	Enabled bool
	Start   time.Time
	End     time.Time
}

// UIState stores state for time window controls.
type UIState struct {
	Open       bool
	Focus      int
	StartInput textinput.Model
	EndInput   textinput.Model
	ErrorMsg   string
	DraftStart time.Time
	DraftEnd   time.Time
	OrigWindow Window
	Step       time.Duration
}

func InitInput(layout string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = layout
	ti.CharLimit = len(layout)
	ti.Width = len(layout)
	ti.Prompt = ""
	return ti
}

func NormalizeStep(step time.Duration) time.Duration {
	if step <= 0 {
		return StepDefault
	}
	if step < StepMin {
		return StepMin
	}
	if step > StepMax {
		return StepMax
	}
	return step
}
