package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/balu-/jsonrpc"
	"github.com/google/uuid"
	"github.com/notnil/chess"
	"golang.org/x/net/websocket"
)

const defaultKillTimeout = time.Hour

var sanDecoder chess.AlgebraicNotation

type SessionMgr struct {
	KillTimeout time.Duration
	sessions    sync.Map
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
		roomID := "custom-" + uuid.NewString()
		sess, loaded := mgr.sessions.LoadOrStore(roomID, &Session{})
		if !loaded {
			sess := sess.(*Session)
			sess.init(roomID, mgr, customFEN)
			return roomID, nil
		}
	}
}

func (mgr *SessionMgr) killSession(roomID string) {
	mgr.sessions.Delete(roomID)
}

type Session struct {
	mtx         sync.Mutex
	roomID      string
	game        *chess.Game
	whiteMove   [2]string
	blackMove   [2]string
	clients     []*jsonrpc.JsonRPC
	killTimer   *time.Timer
	killTimeout time.Duration
}

func (sess *Session) init(roomID string, mgr *SessionMgr, options ...func(*chess.Game)) {
	log.Printf("[new session: %s]", roomID)

	sess.roomID = roomID
	sess.game = chess.NewGame(options...)
	sess.killTimeout = mgr.KillTimeout
	if sess.killTimeout == 0 {
		sess.killTimeout = defaultKillTimeout
	}
	sess.killTimer = time.NewTimer(sess.killTimeout)

	go func() {
		<-sess.killTimer.C
		mgr.killSession(sess.roomID)
		log.Printf("[session expired: %s]", roomID)
	}()
}

func (sess *Session) getUpdate() *Update {
	update := &Update{
		FEN:       sess.game.FEN(),
		WhiteMove: sess.whiteMove,
		BlackMove: sess.blackMove,
	}
	switch sess.game.Outcome() {
	case chess.NoOutcome:
	case chess.Draw:
		update.Message = "Draw"
	default:
		update.Message = "Checkmate"
	}
	return update
}

func (sess *Session) Move(san string, resp *bool) error {
	log.Printf("[%s] %s", sess.roomID, san)

	sess.mtx.Lock()
	defer sess.mtx.Unlock()

	move, err := sanDecoder.Decode(sess.game.Position(), san)
	if err != nil {
		return err
	}
	if err := sess.game.Move(move); err != nil {
		*resp = false
		return nil
	}
	if sess.game.Position().Board().Piece(move.S2()).Color() == chess.White {
		sess.whiteMove[0] = move.S1().String()
		sess.whiteMove[1] = move.S2().String()
	} else {
		sess.blackMove[0] = move.S1().String()
		sess.blackMove[1] = move.S2().String()
	}

	update := sess.getUpdate()
	for _, client := range sess.clients {
		sess.updateClient(client, update)
	}

	*resp = true
	return nil
}

func (sess *Session) addClient(client *jsonrpc.JsonRPC) {
	sess.mtx.Lock()
	defer sess.mtx.Unlock()
	sess.clients = append(sess.clients, client)
	sess.killTimer.Stop()
}

func (sess *Session) removeClient(client *jsonrpc.JsonRPC) {
	sess.mtx.Lock()
	defer sess.mtx.Unlock()
	if len(sess.clients) == 1 {
		sess.clients = nil
		sess.killTimer.Reset(sess.killTimeout)
		return
	}
	for i, cl := range sess.clients {
		if cl == client {
			sess.clients = append(sess.clients[:i], sess.clients[i+1:]...)
			return
		}
	}
}

func (sess *Session) updateClient(client *jsonrpc.JsonRPC, update *Update) {
	unused := false
	client.Call("Session.Update", update, &unused)
}

func (sess *Session) serve(ws *websocket.Conn) {
	client := jsonrpc.NewJsonRpc(ws)
	client.Register(sess, "")

	sess.addClient(client)

	go func() {
		<-time.NewTimer(time.Second / 2).C // artificial delay just to show fancy loader
		sess.updateClient(client, sess.getUpdate())
	}()
	client.Serve()

	sess.removeClient(client)
}

type Update struct {
	FEN       string    `json:"fen"`
	WhiteMove [2]string `json:"wm"`
	BlackMove [2]string `json:"bm"`
	Message   string    `json:"msg"`
}
