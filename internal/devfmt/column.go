package devfmt

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
	case "value", "raw_value":
		return RolePrimary
	case "group_category", "category", "id", "field_label", "field_name":
		return RoleSecondary
	default:
		return RoleNormal
	}
}

func devfmtColumnSchema() siftly.ColumnSchema {
	return siftly.ColumnSchema{
		DefaultMinWidth: 8,
		DefaultWeight:   1.0,
		RoleForName:     detectRole,
		RoleDefaults: map[ui.ColumnRole]siftly.RoleLayout{
			RolePrimary: {
				MinWidth: 40,
				Weight:   5.0,
			},
			RoleSecondary: {
				MinWidth: 12,
				Weight:   2.0,
			},
		},
		ColumnDefaults: map[string]siftly.RoleLayout{
			"group_category": {MinWidth: 14, Weight: 1.0},
			"category":       {MinWidth: 18, Weight: 1.0},
			"id":             {MinWidth: 20, Weight: 2.0},
			"field_name":     {MinWidth: 24, Weight: 2.0},
			"field_label":    {MinWidth: 24, Weight: 2.0},
			"value":          {MinWidth: 48, Weight: 6.0},
			"raw_value":      {MinWidth: 48, Weight: 4.0},
			"time":           {MinWidth: 20, Weight: 1.0},
			"agentid":        {MinWidth: 12, Weight: 1.0},
			"extra":          {MinWidth: 20, Weight: 1.0},
			"desc":           {MinWidth: 30, Weight: 2.0},
		},
	}
}
