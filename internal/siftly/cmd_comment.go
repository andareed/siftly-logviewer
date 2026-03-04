package siftly

import (
	"github.com/andareed/siftly-hostlog/internal/shared/logging"
)

func (m *Model) addComment(comment string) {
	logging.Debug("CommentCurrent called..")
	if (m.cursor) < 0 || m.cursor >= len(m.table.filteredIndices) {
		return
	}

	idx := m.table.filteredIndices[m.cursor]
	hashId := m.table.rows[idx].ID
	if comment == "" {
		delete(m.table.commentRows, hashId)
		logging.Infof("Clear comment Index[%d] on HashID[%d]", idx, hashId)
		return
		//TODO: Probably need this sending a notificatoin
	}
	m.table.commentRows[hashId] = comment
	logging.Infof("Setting Comment[%s] to Index[%d] on HashID[%d]", comment, idx, hashId)
}

func (m *Model) getCommentContent(rowIdx uint64) string {
	// Probably want some error checking around the rowIdx
	if c, ok := m.table.commentRows[rowIdx]; ok && c != "" {
		return c
	}
	return "" // No comment, so returning blank
}

func (m *Model) refreshDrawerContent() {
	logging.Debug("refreshDrawerContent called..")
	currentComment := m.getCommentContent(m.currentRowHashID())
	logging.Debugf("Comment Input and Drawer Port being set to: %s", currentComment)
	m.drawerPort.SetContent(currentComment)
}
