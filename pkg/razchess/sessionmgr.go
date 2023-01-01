package razchess

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/notnil/chess"
	"golang.org/x/net/websocket"
)

const DefaultKillTimeout = time.Hour

type SessionMgr struct {
	killTimeout time.Duration
	sessions    sync.Map
	db          *DB
}

func NewSessionMgr(redisURL string, killTimeout time.Duration) *SessionMgr {
	if killTimeout == 0 {
		killTimeout = DefaultKillTimeout
	}
	mgr := &SessionMgr{
		killTimeout: killTimeout,
	}
	if len(redisURL) > 0 {
		db, err := NewDB(redisURL)
		if err != nil {
			log.Println("Redis error:", err)
		} else {
			for roomID, fen := range db.LoadSessions() {
				customFEN, err := chess.FEN(fen)
				if err != nil {
					log.Println("FEN error:", err)
					continue
				}
				sess := &Session{}
				sess.init(roomID, mgr, customFEN)
				mgr.sessions.Store(roomID, sess)
				log.Printf("[Session loaded from persistent storage: %s]", roomID)
			}
			mgr.db = db
		}
	}
	return mgr
}

func (mgr *SessionMgr) GetSession(roomID string) *Session {
	sess, loaded := mgr.sessions.LoadOrStore(roomID, &Session{})
	if !loaded {
		sess.(*Session).init(roomID, mgr)
	}
	return sess.(*Session)
}

func (mgr *SessionMgr) GetSessionServer(roomID string) http.Handler {
	return websocket.Handler(mgr.GetSession(roomID).serve)
}

func (mgr *SessionMgr) NewCustomSession(fen string) (string, error) {
	customFEN, err := chess.FEN(fen)
	if err != nil {
		return "", err
	}
	for {
		roomID := "custom-" + GenerateID(6)
		sess, loaded := mgr.sessions.LoadOrStore(roomID, &Session{})
		if !loaded {
			sess := sess.(*Session)
			sess.init(roomID, mgr, customFEN)
			return roomID, nil
		}
	}
}

func (mgr *SessionMgr) updateSession(roomID, fen string) {
	if mgr.db != nil {
		mgr.db.SaveSession(roomID, fen, mgr.killTimeout)
	}
}

func (mgr *SessionMgr) killSession(roomID string) {
	mgr.sessions.Delete(roomID)
}
