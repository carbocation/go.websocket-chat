/*
Connection tools.

Note that these depend on the existence of a package-global variable 'h'
that is the hub of all connections.

*/
package main

import (
	"fmt"
	"math/rand"
	"time"

	"code.google.com/p/go.net/websocket"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan string
}

//Sends a message to the user at the other end of this websocket connection
func (c *connection) Send(message string) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	fmt.Printf("msg: %+v connection: %v\n", message, c)

	select {
	//If the message is sent, move onto the next connected user
	case c.send <- message:

	//If this is getting called but there is no message to send, implied disconnect
	default:
		//Tell the hub to unregister us, close the send channel, and close the websocket
		h.unregister <- c
	}
}

//Send messages
func (c *connection) reader() {
	//Shouldn't need to c.ws.Close() here because ultimately
	// this will cause the deferred unregister in wsHandler() to fire
	//defer c.ws.Close()
	for {
		var message string
		err := websocket.Message.Receive(c.ws, &message)
		if err != nil {
			//There will be no message to send to the hub
			break
		}

		//Send the message to the hub
		h.broadcast <- message
	}
}

//Receive messages that were sent
func (c *connection) writer() {
	//Shouldn't need to c.ws.Close() here because ultimately
	// this will cause the deferred unregister in wsHandler() to fire
	//defer c.ws.Close()
	for message := range c.send {
		err := websocket.Message.Send(c.ws, message)
		if err != nil {
			break
		}
	}
}

func wsHandler(ws *websocket.Conn) {
	//Buffer up to 256 messages for this client
	c := &connection{send: make(chan string, 256), ws: ws}

	//Refers to the (global var) hub
	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	c.reader()
}
