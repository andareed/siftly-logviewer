package ui

type ColumnRole int

const (
	RoleNormal ColumnRole = iota
	RolePrimary
	RoleSecondary
)

type ColumnMeta struct {
	Name     string
	Index    int
	Role     ColumnRole
	Visible  bool
	MinWidth int
	Weight   float64
	Width    int
}

func LayoutColumns(cols []ColumnMeta, totalWidth int) []ColumnMeta {
	if totalWidth <= 0 {
		return cols
	}

	minSum := 0
	weightSum := 0.0
	for i := range cols {
		if !cols[i].Visible {
			continue
		}
		minSum += cols[i].MinWidth
		weightSum += cols[i].Weight
	}

	if minSum >= totalWidth {
		for i := range cols {
			if !cols[i].Visible {
				continue
			}
			if cols[i].MinWidth > totalWidth {
				cols[i].Width = totalWidth
			} else {
				cols[i].Width = cols[i].MinWidth
			}
		}
		return cols
	}

	remaining := totalWidth - minSum
	for i := range cols {
		if !cols[i].Visible {
			cols[i].Width = 0
			continue
		}
		extra := 0
		if weightSum > 0 {
			extra = int(float64(remaining) * (cols[i].Weight / weightSum))
		}
		cols[i].Width = cols[i].MinWidth + extra
	}

	return cols
}
