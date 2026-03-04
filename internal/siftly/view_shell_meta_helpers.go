package siftly

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func plainMetaFields(fields []metaField) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, field.label+" "+field.value)
	}
	return strings.Join(parts, "  ")
}

func renderStyledMetaFields(fields []metaField, labelStyle lipgloss.Style, valueStyle lipgloss.Style) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		parts = append(parts, labelStyle.Render(field.label)+" "+valueStyle.Render(field.value))
	}
	return strings.Join(parts, "  ")
}

func truncateEndRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}

func truncateFilenameMiddlePreserveExt(name string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(name)
	if len(r) <= max {
		return name
	}

	ext := filepath.Ext(name)
	if ext == "" {
		return truncateMiddleRunes(name, max)
	}

	extRunes := []rune(ext)
	if len(extRunes) >= max {
		return string(extRunes[len(extRunes)-max:])
	}

	base := strings.TrimSuffix(name, ext)
	baseRunes := []rune(base)
	roomForBase := max - len(extRunes)
	if roomForBase <= 0 {
		return string(extRunes[len(extRunes)-max:])
	}
	if len(baseRunes) <= roomForBase {
		return base + ext
	}
	if roomForBase == 1 {
		return "…" + ext
	}

	core := roomForBase - 1 // reserve one rune for ellipsis
	tail := 4
	if half := core / 2; tail > half {
		tail = half
	}
	if tail < 1 {
		tail = 1
	}
	head := core - tail
	if head < 1 {
		head = 1
		tail = core - head
	}
	if tail > len(baseRunes)-head {
		tail = len(baseRunes) - head
	}
	return string(baseRunes[:head]) + "…" + string(baseRunes[len(baseRunes)-tail:]) + ext
}

func truncateMiddleRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	core := max - 1
	head := (core + 1) / 2
	tail := core - head
	return string(r[:head]) + "…" + string(r[len(r)-tail:])
}
