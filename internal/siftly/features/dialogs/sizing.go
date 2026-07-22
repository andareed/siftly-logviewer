package dialogs

const (
	dialogHorizontalMargin = 4
	dialogVerticalMargin   = 4
	minimumDialogWidth     = 6
)

func responsiveDialogWidth(terminalWidth, preferred, minimum int) int {
	if preferred < minimum {
		preferred = minimum
	}
	if terminalWidth <= 0 {
		return preferred
	}
	available := terminalWidth - dialogHorizontalMargin
	if available < minimumDialogWidth {
		available = terminalWidth
	}
	if available < minimumDialogWidth {
		available = minimumDialogWidth
	}
	if preferred > available {
		return available
	}
	return preferred
}

func responsiveDialogHeight(terminalHeight, preferred int) int {
	if terminalHeight <= 0 {
		return preferred
	}
	margin := dialogVerticalMargin
	if terminalHeight < 16 {
		margin = 2
	}
	available := terminalHeight - margin
	if available < 4 {
		available = terminalHeight
	}
	if available < 2 {
		available = 2
	}
	if preferred > 0 && preferred < available {
		return preferred
	}
	return available
}

func boundedListRows(heightLimit, fixedOuterRows, contentRows, maximum int) int {
	rows := heightLimit - fixedOuterRows
	if rows < 0 {
		rows = 0
	}
	if contentRows >= 0 && rows > contentRows {
		rows = contentRows
	}
	if maximum > 0 && rows > maximum {
		rows = maximum
	}
	return rows
}

func clampDialogWidth(width, minimum, maximum int) int {
	if width < minimum {
		return minimum
	}
	if maximum > 0 && width > maximum {
		return maximum
	}
	return width
}
