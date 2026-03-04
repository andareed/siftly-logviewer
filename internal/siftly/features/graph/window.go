package graph

type Window struct {
	Open    bool
	Height  int
	MaxKeys int
}

func (w Window) HeightOrDefault() int {
	if w.Height <= 0 {
		return 16
	}
	return w.Height
}

func (w Window) MaxKeysOrDefault() int {
	if w.MaxKeys <= 0 {
		return 8
	}
	return w.MaxKeys
}
