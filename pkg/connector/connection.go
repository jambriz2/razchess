package connector

import (
	"io"
	"strings"
	"sync/atomic"

	"github.com/razzie/jsonrpc"
	"github.com/razzie/razchess/pkg/razchess"
	"golang.org/x/net/websocket"
)

type Connection struct {
	ws      io.Closer
	client  *jsonrpc.JsonRPC
	updates chan *razchess.Update
	C       <-chan *razchess.Update
	State   atomic.Pointer[razchess.Update]
	Viewers atomic.Int32
}

func NewConnection(sessionURL string) (*Connection, error) {
	wsURL := strings.NewReplacer("http://", "ws://", "https://", "wss://", "/room/", "/ws/").Replace(sessionURL)
	ws, err := websocket.Dial(wsURL, "", wsURL)
	if err != nil {
		return nil, err
	}
	conn := &Connection{
		ws:      ws,
		client:  jsonrpc.NewJsonRpc(ws),
		updates: make(chan *razchess.Update),
	}
	conn.C = conn.updates
	conn.client.Register(&Session{conn: conn}, "")
	go conn.client.Serve()
	return conn, nil
}

func (conn *Connection) Move(move string) (valid bool) {
	conn.client.Call("Session.Move", move, &valid)
	return
}

func (conn *Connection) Resign(color string) {
	conn.client.Notify("Session.Resign", color)
}

func (conn *Connection) Close() error {
	return conn.ws.Close()
}

func (conn *Connection) update(update *razchess.Update) {
	conn.State.Store(update)
	go func() {
		conn.updates <- update
	}()
}

func (conn *Connection) updateViewCount(count int32) {
	conn.Viewers.Store(count)
}
