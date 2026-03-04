package hostlog

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
	case "details":
		return RolePrimary
	case "id", "time":
		return RoleSecondary
	default:
		return RoleNormal
	}
}

func hostlogColumnSchema() siftly.ColumnSchema {
	return siftly.ColumnSchema{
		DefaultMinWidth: 8,
		DefaultWeight:   1.0,
		RoleForName:     detectRole,
		RoleDefaults: map[ui.ColumnRole]siftly.RoleLayout{
			RolePrimary: {
				MinWidth: 30,
				Weight:   5.0,
			},
			RoleSecondary: {
				MinWidth: 12,
				Weight:   2.0,
			},
		},
	}
}
