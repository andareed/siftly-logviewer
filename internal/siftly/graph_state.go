package siftly

// bumpGraphDataVersion invalidates cached graph series data whenever the
// filtered dataset/order changes.
func (m *Model) bumpGraphDataVersion() {
	m.view.graphDataVersion++
	m.view.graphCache.valid = false
}
