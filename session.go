package main

import (
	"log"
	"sync"
	"time"

	"github.com/balu-/jsonrpc"
	"github.com/notnil/chess"
	"golang.org/x/net/websocket"
)

const killTimeout = time.Hour

var sessions sync.Map
var sanDecoder chess.AlgebraicNotation

func GetSession(roomID string) *Session {
	sess, loaded := sessions.LoadOrStore(roomID, &Session{})
	if !loaded {
		sess.(*Session).init(roomID)
	}
	return sess.(*Session)
}

type Session struct {
	mtx       sync.Mutex
	roomID    string
	game      *chess.Game
	WhiteMove [2]string
	BlackMove [2]string
	clients   []*jsonrpc.JsonRPC
	killTimer *time.Timer
}

func (sess *Session) init(roomID string) {
	log.Printf("[new session: %s]", roomID)

	sess.roomID = roomID
	sess.game = chess.NewGame()
	sess.killTimer = time.NewTimer(killTimeout)
	sess.killTimer.Stop()

	go func() {
		<-sess.killTimer.C
		sessions.Delete(sess.roomID)
		log.Printf("[session expired: %s]", roomID)
	}()
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

	sess.WhiteMove[0] = move.S1().String()
	sess.WhiteMove[1] = move.S2().String()

	/*update := &Update{
		FEN:       sess.game.FEN(),
		WhiteMove: sess.WhiteMove,
		BlackMove: sess.BlackMove,
	}*/

	fen := sess.game.FEN()
	unused := false
	for _, client := range sess.clients {
		client.Call("Session.Update", fen, &unused)
	}

	*resp = true
	return nil
}

func (sess *Session) serve(ws *websocket.Conn) {
	client := jsonrpc.NewJsonRpc(ws)
	client.Register(sess, "")

	<-time.NewTimer(time.Second).C // artificial delay just to show fancy loader

	sess.mtx.Lock()
	sess.killTimer.Stop()
	sess.clients = append(sess.clients, client)
	sess.mtx.Unlock()

	var resp bool
	go client.Call("Session.Update", sess.game.FEN(), &resp)
	client.Serve()

	sess.mtx.Lock()
	defer sess.mtx.Unlock()
	if len(sess.clients) == 1 {
		sess.clients = nil
		sess.killTimer.Reset(killTimeout)
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
	FEN       string `json:"fen"`
	WhiteMove [2]string
	BlackMove [2]string
}
