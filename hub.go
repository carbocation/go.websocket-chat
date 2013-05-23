package main

import (
	"fmt"
	"sync"
)

type connectionMap struct {
	m  map[*connection]struct{}
	mu sync.RWMutex
	//exists bool
}

type hub struct {
	// Registered connections.
	connections connectionMap

	// Inbound messages from the connections.
	//The buffer, if any, guarantees the number of
	//messages which will be received by every client in order
	broadcast chan string

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

func NewHub() *hub {
	return &hub{
		broadcast:   make(chan string, 256), //Guarantee up to 256 messages in order
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

			//If not a goroutine messages will go in order (unless there is a goroutine internally)
			//If a goroutine, no guarantee about message order
			h.bcast(message)
		}
	}
}

func (h *hub) connect(connection *connection) {
	h.connections.mu.Lock()
	h.connections.m[connection] = struct{}{}
	numCons := len(h.connections.m)
	h.connections.mu.Unlock()

	//Unless register and unregister have a buffer, make sure any messaging during these
	//processes is concurrent.
	go func() { h.broadcast <- fmt.Sprintf("hub.connect: %v connected", connection) }()
	fmt.Printf("hub.connect: %v connected\n", connection)
	fmt.Printf("hub.connect: %d clients currently connected\n", numCons)
}

func (h *hub) disconnect(connection *connection) {
	h.connections.mu.Lock()
	delete(h.connections.m, connection)
	numCons := len(h.connections.m)
	h.connections.mu.Unlock()

	connection.mu.Lock()
	connection.dead = true
	connection.mu.Unlock()

	close(connection.send)
	connection.ws.Close()

	//Unless register and unregister have a buffer, make sure any messaging during these
	//processes is concurrent.
	go func() { h.broadcast <- fmt.Sprintf("hub.disconnect: %v disconnected", connection) }()
	fmt.Printf("\nhub.disconnect: FINAL NOTICE %v disconnected FINAL NOTICE\n", connection)
	fmt.Printf("hub.connect: %d clients currently connected\n", numCons)
}

func (h *hub) bcast(message string) {
	//RLock here would guarantee that the map won't change while we iterate over it BUT other goroutines
	// could read the next message simultaneously, so message order is not guaranteed. However, concurrency
	// is maximized.
	//Lock here would guarantee that the map won't change while we iterate over it AND that
	// this is the only goroutine currently reading the map (i.e., it would preserve message order). The
	// degree to which concurrency is impaired depends on whether conn.Send() is called as a goroutine or not.
	//If conn.Send() is called as a goroutine, then choosing between Lock or RLock is of minimal importance,
	// as they would both protect the map just until each connection was launched (but not finished).
	//If conn.Send() is called as a normal routine, then
	h.connections.mu.RLock()

	//Count launched routines
	i := 0
	finChan := make(chan struct{})
	for conn := range h.connections.m {
		//For every connected user, do something with the message or disconnect
		//Each user may have a different delay, but no user blocks others

		//To simulate different users getting different messages, we'll send timestamps and sleep, too:
		//TODO TEST THIS with/wo the go
		fmt.Printf("hub.bcast: conn.Send'ing message '''%v''' to conn %v\n", message, conn)

		//If this is a goroutine, then mutex
		go conn.Send(message, finChan)
		i++
	}

	//Done iterating over the map
	h.connections.mu.RUnlock()

	//Drain all finChan values; afterwards, we'll unblock
	for i > 0 {
		select {
		case <-finChan:
			i--
		}
	}

	fmt.Printf("hub.bcast: bcast'ing message ```%v``` is done.", message)
}
