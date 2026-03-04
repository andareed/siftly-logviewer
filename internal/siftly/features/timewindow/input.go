package timewindow

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidStart  = errors.New("invalid start time")
	ErrInvalidEnd    = errors.New("invalid end time")
	ErrStartAfterEnd = errors.New("start is after end")
)

func SyncDraftFromInputs(ui *UIState, hasBounds bool, loc *time.Location) {
	if ui == nil || !hasBounds {
		return
	}
	startStr := strings.TrimSpace(ui.StartInput.Value())
	endStr := strings.TrimSpace(ui.EndInput.Value())
	start, err := parseInputTime(startStr, loc)
	if err == nil {
		ui.DraftStart = start
	}
	end, err := parseInputTime(endStr, loc)
	if err == nil {
		ui.DraftEnd = end
	}
}

func ParseInputWindow(startInput, endInput string, loc *time.Location, min, max time.Time) (Window, error) {
	start, err := parseInputTime(strings.TrimSpace(startInput), loc)
	if err != nil {
		return Window{}, ErrInvalidStart
	}
	end, err := parseInputTime(strings.TrimSpace(endInput), loc)
	if err != nil {
		return Window{}, ErrInvalidEnd
	}
	if start.After(end) {
		return Window{}, ErrStartAfterEnd
	}

	start = Clamp(start, min, max)
	end = Clamp(end, min, max)
	if start.After(end) {
		return Window{}, ErrStartAfterEnd
	}

	return Window{
		Enabled: true,
		Start:   start,
		End:     end,
	}, nil
}

func parseInputTime(input string, loc *time.Location) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, errors.New("empty time input")
	}
	if ts, err := time.ParseInLocation(InputLayout, input, loc); err == nil {
		return ts, nil
	}
	if ts, err := time.ParseInLocation(InputLayoutNoSeconds, input, loc); err == nil {
		return ts, nil
	}
	return time.Time{}, errors.New("invalid time input")
}
