package ui

type LayoutInput struct {
	TerminalHeight   int
	TerminalWidth    int
	DrawerOpen       bool
	DrawerContentH   int
	GraphOpen        bool
	GraphHeight      int
	TimeWindowOpen   bool
	TimeWindowHeight int
}

type LayoutOutput struct {
	ViewportHeight int
	ViewportWidth  int
	DrawerHeight   int
}

func ComputeLayout(in LayoutInput) LayoutOutput {
	height := in.TerminalHeight - 6
	width := in.TerminalWidth - 6
	out := LayoutOutput{
		ViewportHeight: height,
		ViewportWidth:  width,
	}
	if in.DrawerOpen {
		out.DrawerHeight = in.DrawerContentH + 2
		out.ViewportHeight -= out.DrawerHeight
	}
	if in.GraphOpen {
		out.ViewportHeight -= in.GraphHeight + 2
	}
	if in.TimeWindowOpen {
		out.ViewportHeight -= in.TimeWindowHeight
	}
	if out.ViewportHeight < 0 {
		out.ViewportHeight = 0
	}
	if out.ViewportWidth < 0 {
		out.ViewportWidth = 0
	}
	return out
}
