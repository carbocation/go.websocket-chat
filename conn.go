/*
Connection tools.

Note that these depend on the existence of a package-global variable 'h'
that is the hub of all connections.

*/
package main

import (
	"fmt"
	"sync"

	"code.google.com/p/go.net/websocket"
)

type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	//If the buffer is reached, the client will be
	//considered to have timed out and disconnected.
	//This can really only happen if message order is not preserved.
	send chan string

	//Have we received a kill signal?
	dead bool

	mu sync.RWMutex
}

//Sends a message to the user at the other end of this websocket connection
//Notify the hub when finished by sending an empty struct over the fin channel
func (c *connection) Send(message string, fin chan struct{}, h *hub) {
	defer func() {
		fmt.Printf("conn.Send: message '''%s''' to %v\n", message, c)

		//Tell the calling function that this goroutine is done sending
		fin <- struct{}{}
	}()

	c.mu.RLock()
	if c.dead {
		//Channel is already dead, we cannot send on it anymore and we must exit
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	//We don't want to try to send over the channel if another
	//goroutine has closed this channel in the meantime. Thus, we
	//must block writing before we send over this channel.
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
func (c *connection) reader(h *hub) {
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
