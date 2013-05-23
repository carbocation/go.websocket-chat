/*
Connection tools.

Note that these depend on the existence of a package-global variable 'h'
that is the hub of all connections.

*/
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"code.google.com/p/go.net/websocket"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan string

	//Have we received a kill signal?
	dead bool

	mu sync.RWMutex
}

func randDelay() time.Duration {
	return time.Duration(rand.Intn(10)) * time.Second
}

//Sends a message to the user at the other end of this websocket connection
func (c *connection) Send(message string) {
	delay := randDelay()
	fmt.Printf("conn.Send: message '''%s''' will be sent to connection %v after: %v\n", message, c, delay)
	time.Sleep(delay)
	defer fmt.Printf("conn.Send: message '''%s''' to %v happened after %v\n", message, c, delay)

	c.mu.RLock()
	if c.dead {
		//Channel is already dead, we cannot send on it anymore and we must exit
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	c.mu.Lock()
	select {
	//If the message is sent over the websocket, unlock this connection and continue
	case c.send <- message:
		c.mu.Unlock()

	//If we cannot send, this means that the user's buffer is full. At this point we basically
	//assume that the user disconnected or is just stuck.
	default:
		//Tell the hub to unregister us, close the send channel, and close the websocket
		fmt.Printf("conn.Send: Implied disconnect of %+v\n", c)
		//Unlock before unregistering since the act of unregistering triggers changes in c
		c.mu.Unlock()
		h.unregister <- c
	}
}

//Send messages for broadcasting
func (c *connection) reader() {
	//Shouldn't need to c.ws.Close() here because ultimately
	// this will cause the deferred unregister in wsHandler() to fire
	//defer c.ws.Close()
	defer fmt.Printf("conn.reader: reader for %+v exited\n", c)
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

//Messages that were broadcast to this particular connection
//go down the wire to the user here.
func (c *connection) writer() {
	//Shouldn't need to c.ws.Close() here because ultimately
	// this will cause the deferred unregister in wsHandler() to fire
	//defer c.ws.Close()
	defer fmt.Printf("conn.writer: writer for %+v exited\n", c)
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

	//Register this connection into the (global var) hub
	h.register <- c

	//If this deferred function gets called, it implies that
	// writer and reader already exited
	defer func() {
		h.unregister <- c
		fmt.Printf("conn.wsHandler: SOCKET CLOSED %v is completely closed SOCKET CLOSED\n", c)
	}()
	go c.writer()
	c.reader()
}
