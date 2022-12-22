package main

import (
	"log"
	"sync"
	"time"

	"github.com/balu-/jsonrpc"
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

func (sess *Session) init(roomID string, mgr *SessionMgr) {
	log.Printf("[new session: %s]", roomID)

	sess.roomID = roomID
	sess.game = chess.NewGame()
	sess.killTimeout = mgr.KillTimeout
	if sess.killTimeout == 0 {
		sess.killTimeout = defaultKillTimeout
	}
	sess.killTimer = time.NewTimer(sess.killTimeout)
	sess.killTimer.Stop()

	go func() {
		<-sess.killTimer.C
		mgr.killSession(sess.roomID)
		log.Printf("[session expired: %s]", roomID)
	}()
}

func (sess *Session) getUpdate() *Update {
	return &Update{
		FEN:       sess.game.FEN(),
		WhiteMove: sess.whiteMove,
		BlackMove: sess.blackMove,
	}
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
	unused := false
	switch sess.game.Outcome() {
	case chess.NoOutcome:
	case chess.Draw:
		update.Message = "Draw"
	default:
		update.Message = "Checkmate"
	}
	for _, client := range sess.clients {
		client.Call("Session.Update", update, &unused)
	}

	*resp = true
	return nil
}

func (sess *Session) serve(ws *websocket.Conn) {
	client := jsonrpc.NewJsonRpc(ws)
	client.Register(sess, "")

	<-time.NewTimer(time.Second / 2).C // artificial delay just to show fancy loader

	sess.mtx.Lock()
	sess.killTimer.Stop()
	sess.clients = append(sess.clients, client)
	sess.mtx.Unlock()

	var resp bool
	go client.Call("Session.Update", sess.getUpdate(), &resp)
	client.Serve()

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

type Update struct {
	FEN       string    `json:"fen"`
	WhiteMove [2]string `json:"wm"`
	BlackMove [2]string `json:"bm"`
	Message   string    `json:"msg"`
}
