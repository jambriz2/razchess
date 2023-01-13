package razchess

import (
	"sync"
	"time"

	"github.com/notnil/chess"
	"github.com/razzie/jsonrpc"
	"golang.org/x/net/websocket"
)

type Session struct {
	slc      *sessionLifecycle
	mtx      sync.Mutex
	game     *chess.Game
	isCustom bool
	clients  []*jsonrpc.JsonRPC
}

func newSession(slc *sessionLifecycle, game string) (*Session, error) {
	sess := &Session{}
	if err := sess.init(slc, game); err != nil {
		return nil, err
	}
	return sess, nil
}

func (sess *Session) init(slc *sessionLifecycle, game string) error {
	opts, isCustom, err := parseGame(game)
	if err != nil {
		return err
	}
	sess.slc = slc
	sess.game = chess.NewGame(opts...)
	sess.isCustom = isCustom
	return nil
}

// Session.Move is the only exposed RPC function
func (sess *Session) Move(san string, validMove *bool) error {
	sess.mtx.Lock()
	defer sess.mtx.Unlock()

	*validMove = sess.handleMoveSAN(san)
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
		sess.handleMove(validNextMoves[0])
		sess.updateClients()
	}

	go sess.slc.update(sess.gameToString())

	return nil
}

func (sess *Session) handleMove(move *chess.Move) bool {
	if err := sess.game.Move(move); err != nil {
		return false
	}
	return true
}

func (sess *Session) handleMoveSAN(san string) bool {
	if err := sess.game.MoveStr(san); err != nil {
		return false
	}
	return true
}

func (sess *Session) addClient(client *jsonrpc.JsonRPC) {
	sess.mtx.Lock()
	defer sess.mtx.Unlock()
	sess.clients = append(sess.clients, client)
	sess.slc.stopTimer()
}

func (sess *Session) removeClient(client *jsonrpc.JsonRPC) {
	sess.mtx.Lock()
	defer sess.mtx.Unlock()
	if len(sess.clients) == 1 {
		sess.clients = nil
		sess.slc.startTimer()
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
	return newUpdate(sess.game)
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

func (sess *Session) gameToString() string {
	return gameToString(sess.game, sess.isCustom)
}

func (sess *Session) serve(ws *websocket.Conn) {
	client := jsonrpc.NewJsonRpc(ws)
	client.Register(sess, "")

	sess.addClient(client)

	sess.updateClient(client, sess.getUpdate())
	client.Serve()

	sess.removeClient(client)
}
