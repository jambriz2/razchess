package razchess

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

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
			for roomID, game := range db.LoadSessions() {
				opt, err := parseGame(game)
				if err != nil {
					log.Println("Game parse error:", err)
					continue
				}
				sess := &Session{}
				sess.init(roomID, mgr, opt)
				if strings.HasPrefix(game, "pgn:") {
					sess.isCustom = false
				}
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

func (mgr *SessionMgr) NewCustomSession(game string) (string, error) {
	opt, err := parseGame(game)
	if err != nil {
		return "", err
	}
	for {
		roomID := "custom-" + GenerateID(6)
		sess, loaded := mgr.sessions.LoadOrStore(roomID, &Session{})
		if !loaded {
			sess := sess.(*Session)
			sess.init(roomID, mgr, opt)
			return roomID, nil
		}
	}
}

func (mgr *SessionMgr) updateSession(roomID, game string) {
	if mgr.db != nil {
		mgr.db.SaveSession(roomID, game, mgr.killTimeout)
	}
}

func (mgr *SessionMgr) killSession(roomID string) {
	mgr.sessions.Delete(roomID)
}
