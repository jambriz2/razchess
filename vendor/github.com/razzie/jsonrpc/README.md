bidirectional jsonrpc
=====================

## Examples

### Server
```go
package main

import (
	"log"
	"net/http"

	"github.com/razzie/jsonrpc"
	"golang.org/x/net/websocket"
)

type Service struct{}

// Server side 'Service.Ping' RPC function
func (srv *Service) Ping(msg string, resp *string) error {
	*resp = "Pong: " + msg
	return nil
}

func main() {
	http.Handle("/ws", websocket.Handler(serveClient))
	http.ListenAndServe(":8080", nil)
}

// This function serves a single client
func serveClient(ws *websocket.Conn) {
	client := jsonrpc.NewJsonRpc(ws)

    // Need to register server side RPC functions
	client.Register(new(Service), "")

    // Calling client side 'AddOne' RPC function
	var addOneResult int
	client.Call("AddOne", 99, &addOneResult)
	log.Println("AddOne(99):", addOneResult)

	client.Serve()
}
```

### Client (using [simple-jsonrpc-js](https://github.com/jershell/simple-jsonrpc-js))
```javascript
var jrpc = new simple_jsonrpc();
var socket = new WebSocket((window.location.protocol == 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host + '/ws');

socket.onmessage = function(event) {
    jrpc.messageHandler(event.data);
};

jrpc.toStream = function(_msg){
    socket.send(_msg);
};

// Client side 'AddOne' RPC function
jrpc.on('AddOne', function(num) {
    return num + 1;
});

// Calling server side 'Service.Ping' RPC function
jrpc.call('Service.Ping', ['hello']).then(function(response) {
        alert('Server response: ' + response)
});
```
