package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type NoticeState struct {
	Msg  string
	Type string
	Seq  int
}

type ClearNoticeMsg struct{ ID int }

func NoticeText(msg, kind string) string {
	if msg == "" {
		return ""
	}
	var icon string
	switch kind {
	case "info":
		icon = "ℹ"
	case "success":
		icon = "✓"
	case "warn":
		icon = "!"
	case "error":
		icon = "×"
	default:
		icon = ""
	}
	if icon == "" {
		return msg
	}
	return icon + " " + msg
}

func StartNotice(st *NoticeState, msg, kind string, d time.Duration) tea.Cmd {
	if st == nil {
		return nil
	}
	id := SetNotice(st, msg, kind)
	return tea.Tick(d, func(time.Time) tea.Msg { return ClearNoticeMsg{ID: id} })
}

func SetNotice(st *NoticeState, msg, kind string) int {
	if st == nil {
		return 0
	}
	st.Msg = msg
	st.Type = kind
	st.Seq++
	return st.Seq
}

func ApplyClearNotice(st *NoticeState, msg ClearNoticeMsg) bool {
	if st == nil {
		return false
	}
	if msg.ID != st.Seq {
		return false
	}
	st.Msg = ""
	st.Type = ""
	return true
}

func ClearNotice(st *NoticeState) {
	if st == nil {
		return
	}
	st.Msg = ""
	st.Type = ""
	st.Seq++
}

func (n *NoticeState) Start(msg, kind string, d time.Duration) tea.Cmd {
	return StartNotice(n, msg, kind, d)
}

func (n *NoticeState) Set(msg, kind string) int {
	return SetNotice(n, msg, kind)
}

func (n *NoticeState) ApplyClear(msg ClearNoticeMsg) bool {
	return ApplyClearNotice(n, msg)
}

func (n *NoticeState) Clear() {
	ClearNotice(n)
}
