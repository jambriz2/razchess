package razchess

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
			mgr.db = db
			mgr.loadSessions()
		}
	}
	return mgr
}

func (mgr *SessionMgr) CreateSession(game string) (string, error) {
	slc := newSessionLifecycle(mgr, "")
	sess, err := newSession(slc, game)
	if err != nil {
		return "", err
	}
	for {
		roomID := GenerateID(6)
		if _, loaded := mgr.sessions.LoadOrStore(roomID, sess); !loaded {
			slc.resetRoomID(roomID)
			if len(game) > 0 {
				log.Printf("[new custom session: %s] %s", roomID, game)
			} else {
				log.Printf("[new session: %s]", roomID)
			}
			go slc.update(sess.gameToString())
			return roomID, nil
		}
	}
}

func (mgr *SessionMgr) ServeRPC(w http.ResponseWriter, r *http.Request, roomID string) {
	if len(roomID) == 0 {
		http.Error(w, "Empty roomID", http.StatusBadRequest)
		return
	}
	sess := mgr.getOrCreateSession(roomID)
	websocket.Handler(sess.serve).ServeHTTP(w, r)
}

func (mgr *SessionMgr) MoveHistoryToGIF(w io.Writer, roomID string) error {
	sess, ok := mgr.sessions.Load(roomID)
	if !ok {
		return fmt.Errorf("session not found: %s", roomID)
	}
	moves, positions := sess.(*Session).getMoveHistory()
	return MoveHistoryToGIF(w, moves, positions)
}

func (mgr *SessionMgr) getOrCreateSession(roomID string) *Session {
	sess, loaded := mgr.sessions.LoadOrStore(roomID, &Session{})
	if !loaded {
		sess.(*Session).init(newSessionLifecycle(mgr, roomID), "")
		log.Printf("[new session: %s]", roomID)
	}
	return sess.(*Session)
}

func (mgr *SessionMgr) loadSessions() {
	for roomID, game := range mgr.db.LoadSessions() {
		log.Printf("[Loading session from persistent storage: %s]", roomID)
		sess, err := newSession(newSessionLifecycle(mgr, roomID), game)
		if err != nil {
			log.Println(err)
			continue
		}
		mgr.sessions.Store(roomID, sess)
	}
}

func (mgr *SessionMgr) updateSession(roomID, game string) {
	if mgr.db != nil && len(roomID) > 0 {
		mgr.db.SaveSession(roomID, game, mgr.killTimeout)
	}
}

func (mgr *SessionMgr) killSession(roomID string) {
	log.Printf("[session expired: %s]", roomID)
	mgr.sessions.Delete(roomID)
}
