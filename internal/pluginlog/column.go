package pluginlog

import (
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly"
	"github.com/andareed/siftly-hostlog/internal/siftly/ui"
)

type ColumnRole = ui.ColumnRole

const (
	RoleNormal    = ui.RoleNormal
	RolePrimary   = ui.RolePrimary
	RoleSecondary = ui.RoleSecondary
)

func detectRole(name string) ColumnRole {
	n := strings.ToLower(strings.TrimSpace(name))
	switch n {
	case "message":
		return RolePrimary
	case "time", "process", "module", "function":
		return RoleSecondary
	default:
		return RoleNormal
	}
}

func pluginlogColumnSchema() siftly.ColumnSchema {
	return siftly.ColumnSchema{
		DefaultMinWidth: 8,
		DefaultWeight:   0.0,
		RoleForName:     detectRole,
		RoleDefaults: map[ui.ColumnRole]siftly.RoleLayout{
			RolePrimary: {
				MinWidth: 44,
				Weight:   1.0,
			},
			RoleSecondary: {
				MinWidth: 12,
				Weight:   0.0,
			},
		},
		ColumnDefaults: map[string]siftly.RoleLayout{
			"process": {
				MinWidth: 7,
				Weight:   0.0,
			},
			"pid": {
				MinWidth: 5,
				Weight:   0.0,
			},
			"epoch": {
				MinWidth: 18,
				Weight:   0.0,
			},
			"time": {
				MinWidth: 19,
				Weight:   0.0,
			},
			"level": {
				MinWidth: 8,
				Weight:   0.0,
			},
			"module": {
				MinWidth: 15,
				Weight:   0.0,
			},
			"function": {
				MinWidth: 28,
				Weight:   0.0,
			},
			"line": {
				MinWidth: 6,
				Weight:   0.0,
			},
			"message": {
				MinWidth: 44,
				Weight:   1.0,
			},
		},
	}
}
