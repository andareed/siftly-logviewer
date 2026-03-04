package siftly

import "strings"

func columnsNoticeText(toggled, missing []string) string {
	status := ""
	if len(toggled) > 0 {
		status = "Toggled: " + strings.Join(toggled, ", ")
	}
	if len(missing) > 0 {
		msg := "Unknown: " + strings.Join(missing, ", ")
		if status != "" {
			status += " · " + msg
		} else {
			status = msg
		}
	}
	if status == "" {
		status = "No columns matched"
	}
	return status
}
