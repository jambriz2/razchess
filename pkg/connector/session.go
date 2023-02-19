package connector

import (
	"github.com/razzie/razchess/pkg/razchess"
)

type Session struct {
	conn *Connection
}

func (sess *Session) Update(update *razchess.Update, unused *bool) error {
	sess.conn.update(update)
	return nil
}

func (sess *Session) UpdateViewCount(count int32, unused *bool) error {
	sess.conn.updateViewCount(count)
	return nil
}
