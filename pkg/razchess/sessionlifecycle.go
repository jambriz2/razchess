package razchess

import (
	"time"
)

type sessionLifecycle struct {
	mgr         *SessionMgr
	roomID      string
	killTimer   *time.Timer
	killTimeout time.Duration
}

func newSessionLifecycle(mgr *SessionMgr, roomID string) *sessionLifecycle {
	slc := &sessionLifecycle{
		mgr:         mgr,
		roomID:      roomID,
		killTimer:   time.NewTimer(mgr.killTimeout),
		killTimeout: mgr.killTimeout,
	}
	go func() {
		<-slc.killTimer.C
		mgr.killSession(roomID)
	}()
	return slc
}

func (slc *sessionLifecycle) resetRoomID(roomID string) {
	slc.roomID = roomID
}

func (slc *sessionLifecycle) update(game string) {
	slc.mgr.updateSession(slc.roomID, game)
}

func (slc *sessionLifecycle) startTimer() {
	slc.killTimer.Reset(slc.killTimeout)
}

func (slc *sessionLifecycle) stopTimer() {
	slc.killTimer.Stop()
}
