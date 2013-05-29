/*
Websocket-Hub is heavily derived from Gary Burd's example chat service,
and much of this is copyright Gary Burd, used under the Apache License,
Version 2.0 ( https://github.com/garyburd/go-websocket#license )

The remainder is copyright 2013 James Pirruccello. The entire work is
offered under the Apache License, Version 2.0.
*/
package wshub

import (
	"github.com/garyburd/go-websocket/websocket"
)

//Hub of hubs lives here
var hubs *hubMap = &hubMap{m: map[string]*hub{}}

//Package-global config lives here:
var cfg *config

//TODO(james):
//When registering the handler, pull out the ID that they registered on.
//That will be the channel ID, which determines which hub they register upon.
//Each hub will exist as a separate goroutine. Thus, each page will have a
//different non-blocking hub, but all messages on a given page will be in order.
//TODO(james):
//When sending new data over the socket, look up all parent IDs in an
// ancestry cache; if that's empty then check DB. Finally, send the
// data to anyone registered to any ancestor channels.
//Launch handles all transactions over a websocket connection for a given hub
func Launch(ws *websocket.Conn, id string) {
	//Buffer up to 256 messages for this client
	c := NewConnection(ws, make(chan []byte, 256))

	//Register this connection into the (global var) hub
	h := GetHub(id)
	h.Register(c)

	//If this deferred function gets called, it implies that
	// writer and reader already exited
	defer func() {
		h.Unregister(c)
	}()
	go c.Writer()
	c.Reader(h)
}
