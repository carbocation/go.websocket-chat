package main

import (
	"fmt"
	"math/rand"
	"time"
)

type hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan string

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

func NewHub() *hub {
	return &hub{
		broadcast:   make(chan string),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
	}
}

func (h *hub) run() {
	for {
		select {
		case connection := <-h.register:
			h.connections[connection] = true
		case connection := <-h.unregister:
			delete(h.connections, connection)
			close(connection.send)
		case message := <-h.broadcast:
			//We've received a message that is potentially supposed to be broadcast
			for connection := range h.connections {
				//For every connected user, do something with the message or disconnect

				fmt.Printf("key: %+v", connection)
				//To simulate different users getting different messages, we'll send timestamps and sleep, too:
				time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

				connection.Send(message)
			}
		}
	}
}
