package siftly

func (m *Model) setModeHint(msg string) {
	m.view.modeHintSeq = m.view.notice.Set(msg, "info")
}

func (m *Model) clearModeHint() {
	if m.view.modeHintSeq > 0 && m.view.notice.Seq == m.view.modeHintSeq {
		m.view.notice.Clear()
	}
	m.view.modeHintSeq = 0
}

func (m *Model) setPrefixHint(msg string) {
	m.view.prefixHintSeq = m.view.notice.Set(msg, "info")
}

func (m *Model) clearPrefixHint() {
	if m.view.prefixHintSeq > 0 && m.view.notice.Seq == m.view.prefixHintSeq {
		m.view.notice.Clear()
	}
	m.view.prefixHintSeq = 0
}
