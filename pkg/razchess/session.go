package razchess

import (
	"log"
	"sync"
	"time"

	"github.com/notnil/chess"
	"github.com/razzie/jsonrpc"
	"golang.org/x/net/websocket"
)

var sanDecoder chess.AlgebraicNotation

type Session struct {
	mtx         sync.Mutex
	roomID      string
	game        *chess.Game
	isCustom    bool
	clients     []*jsonrpc.JsonRPC
	killTimer   *time.Timer
	killTimeout time.Duration
	dbUpdater   func()
}

// Session.Move is the only exposed RPC function
func (sess *Session) Move(san string, validMove *bool) error {
	log.Printf("[%s] %s", sess.roomID, san)

	sess.mtx.Lock()
	defer sess.mtx.Unlock()

	*validMove = sess.handleMove(san)
	if !*validMove {
		return nil
	}
	sess.updateClients()

	for i := 0; i < 10; i++ { // limit auto moves
		validNextMoves := sess.game.ValidMoves()
		if len(validNextMoves) != 1 {
			break
		}
		<-time.NewTimer(time.Second / 2).C
		sess.handleMove(sanDecoder.Encode(sess.game.Position(), validNextMoves[0]))
		sess.updateClients()
	}

	go sess.dbUpdater()

	return nil
}

func (sess *Session) init(roomID string, mgr *SessionMgr, options ...func(*chess.Game)) {
	log.Printf("[new session: %s]", roomID)

	sess.roomID = roomID
	sess.game = chess.NewGame(options...)
	sess.isCustom = len(options) > 0
	sess.killTimeout = mgr.killTimeout
	sess.killTimer = time.NewTimer(sess.killTimeout)
	sess.dbUpdater = func() {
		mgr.updateSession(roomID, gameToString(sess.game, sess.isCustom))
	}

	go func() {
		<-sess.killTimer.C
		mgr.killSession(sess.roomID)
		log.Printf("[session expired: %s]", roomID)
	}()
}

func (sess *Session) handleMove(san string) bool {
	move, err := sanDecoder.Decode(sess.game.Position(), san)
	if err != nil {
		return false
	}
	if err := sess.game.Move(move); err != nil {
		return false
	}
	return true
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

func (sess *Session) getUpdate() *Update {
	return newUpdate(sess.game, !sess.isCustom)
}

func (sess *Session) updateClient(client *jsonrpc.JsonRPC, update *Update) {
	client.Notify("Session.Update", update)
}

func (sess *Session) updateClients() {
	update := sess.getUpdate()
	for _, client := range sess.clients {
		sess.updateClient(client, update)
	}
}

func (sess *Session) serve(ws *websocket.Conn) {
	client := jsonrpc.NewJsonRpc(ws)
	client.Register(sess, "")

	sess.addClient(client)

	sess.updateClient(client, sess.getUpdate())
	client.Serve()

	sess.removeClient(client)
}
