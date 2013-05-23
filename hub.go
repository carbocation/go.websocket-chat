package main

import (
	"fmt"
	"sync"
)

type connectionMap struct {
	m  map[*connection]struct{}
	mu sync.RWMutex
}

type hub struct {
	// Registered connections.
	connections connectionMap

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
		connections: connectionMap{m: make(map[*connection]struct{})},
	}
}

func (h *hub) run() {
	for {
		select {
		case connection := <-h.register:
			//Add a connection
			go h.connect(connection)
		case connection := <-h.unregister:
			//Delete a connection
			go h.disconnect(connection)
		case message := <-h.broadcast:
			//We've received a message that is potentially supposed to be broadcast

			//If not a goroutine messages will go in order
			//If a goroutine, no guarantee about message order
			go h.bcast(message)
		}
	}
}

func (h *hub) connect(connection *connection) {
	h.connections.mu.Lock()
	h.connections.m[connection] = struct{}{}
	h.broadcast <- fmt.Sprintf("%p connected", connection)
	fmt.Printf("%v connected\n", connection)
	h.connections.mu.Unlock()
}

func (h *hub) disconnect(connection *connection) {
	h.connections.mu.Lock()
	delete(h.connections.m, connection)
	close(connection.send)
	connection.ws.Close()
	h.broadcast <- fmt.Sprintf("%p disconnected", connection)
	fmt.Printf("%v disconnected\n", connection)
	h.connections.mu.Unlock()
}

func (h *hub) bcast(message string) {
	h.connections.mu.RLock()
	defer h.connections.mu.RUnlock()

	for conn := range h.connections.m {
		//For every connected user, do something with the message or disconnect
		//Each user may have a different delay, but no user blocks others

		//To simulate different users getting different messages, we'll send timestamps and sleep, too:
		conn.Send(message)
	}
}
