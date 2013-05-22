/*
Connection tools.

Note that these depend on the existence of a package-global variable 'h'
that is the hub of all connections.

*/
package main

import (
	"code.google.com/p/go.net/websocket"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan string
}

//Sends a message on the channel.
func (c *connection) Send(message string) {
	select {
	//If the message is sent, move onto the next connected user
	case c.send <- message:
	//If the message fails to send, the user has disconnected
	default:
		delete(h.connections, c)
		close(c.send)
		go c.ws.Close()
	}
}

func (c *connection) reader() {
	for {
		var message string
		err := websocket.Message.Receive(c.ws, &message)
		if err != nil {
			break
		}

		//Sends to the (global var) hub
		h.broadcast <- message
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		err := websocket.Message.Send(c.ws, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func wsHandler(ws *websocket.Conn) {
	c := &connection{send: make(chan string, 256), ws: ws}

	//Refers to the (global var) hub
	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	c.reader()
}
