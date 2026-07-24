package todaylog

import (
	"strconv"
	"strings"
	"time"

	"github.com/andareed/siftly-hostlog/internal/siftly"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

type ColumnRole = ui.ColumnRole
type ColumnMeta = ui.ColumnMeta

const (
	RoleNormal    = ui.RoleNormal
	RolePrimary   = ui.RolePrimary // Key
	RoleSecondary = ui.RoleSecondary
)

func detectRole(name string) ColumnRole {
	n := strings.ToLower(strings.TrimSpace(name))
	switch n {
	case "key":
		return RolePrimary
	case "value", "process", "date":
		return RoleSecondary
	default:
		return RoleNormal
	}
}

func defaultMinWidthForRole(r ColumnRole) int {
	switch r {
	case RolePrimary:
		return 40
	case RoleSecondary:
		return 12
	default:
		return 8
	}
}

func defaultWeightForRole(r ColumnRole) float64 {
	switch r {
	case RolePrimary:
		return 6.0
	case RoleSecondary:
		return 2.0
	default:
		return 1.0
	}
}

func todaylogColumnSchema() siftly.ColumnSchema {
	return siftly.ColumnSchema{
		DefaultMinWidth: 8,
		DefaultWeight:   1.0,
		RoleForName:     detectRole,
		TimeParser:      parseTodaylogUnixSeconds,
		RoleDefaults: map[ui.ColumnRole]siftly.RoleLayout{
			RolePrimary: {
				MinWidth:  40,
				Weight:    6.0,
				WrapLines: 4,
			},
			RoleSecondary: {
				MinWidth: 12,
				Weight:   2.0,
			},
		},
		ColumnDefaults: map[string]siftly.RoleLayout{
			"date": {
				MinWidth: 17,
				Weight:   1.0,
			},
			"timestamp": {
				MinWidth: 12,
				Weight:   0.5,
			},
			"pid": {
				MinWidth: 7,
				Weight:   0.5,
			},
			"process": {
				MinWidth: 12,
				Weight:   1.0,
			},
			"key": {
				MinWidth:  40,
				Weight:    6.0,
				WrapLines: 4,
			},
			"value": {
				MinWidth: 12,
				Weight:   3.0,
			},
		},
	}
}

func parseTodaylogUnixSeconds(cols []string, timeColumnIndex int) (time.Time, bool) {
	if timeColumnIndex < 0 || timeColumnIndex >= len(cols) {
		return time.Time{}, false
	}
	raw := strings.TrimSpace(cols[timeColumnIndex])
	if raw == "" {
		return time.Time{}, false
	}
	secs, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(secs, 0), true
}
