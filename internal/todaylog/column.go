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
	RolePrimary   = ui.RolePrimary // Details/value
	RoleSecondary = ui.RoleSecondary
)

func detectRole(name string) ColumnRole {
	n := strings.ToLower(strings.TrimSpace(name))
	switch n {
	case "value":
		return RolePrimary
	case "key", "process", "date":
		return RoleSecondary
	default:
		return RoleNormal
	}
}

func defaultMinWidthForRole(r ColumnRole) int {
	switch r {
	case RolePrimary:
		return 50
	case RoleSecondary:
		return 12
	default:
		return 8
	}
}

func defaultWeightForRole(r ColumnRole) float64 {
	switch r {
	case RolePrimary:
		return 5.0
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
				MinWidth: 50,
				Weight:   5.0,
			},
			RoleSecondary: {
				MinWidth: 12,
				Weight:   2.0,
			},
		},
		ColumnDefaults: map[string]siftly.RoleLayout{
			"date": {
				MinWidth: 19,
				Weight:   2.0,
			},
			"timestamp": {
				MinWidth: 16,
				Weight:   1.0,
			},
			"pid": {
				MinWidth: 8,
				Weight:   1.0,
			},
			"process": {
				MinWidth: 24,
				Weight:   2.0,
			},
			"key": {
				MinWidth: 24,
				Weight:   3.0,
			},
			"value": {
				MinWidth: 60,
				Weight:   6.0,
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
