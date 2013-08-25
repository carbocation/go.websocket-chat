/*
Websocket-Hub is heavily derived from Gary Burd's example chat service,
and much of this is copyright Gary Burd, used under the Apache License,
Version 2.0 ( https://github.com/garyburd/go-websocket#license )

The remainder is copyright 2013 James Pirruccello. The entire work is
offered under the Apache License, Version 2.0.
*/
package wshub

import (
	"fmt"
	"sync"
)

//hubMap stores all active hubs
type hubMap struct {
	m  map[string](*hub)
	mu sync.RWMutex
}

//BroadcastAll sends a message to every client on every hub
func BroadcastAll(input []byte) {
	hubs.mu.Lock()
	defer hubs.mu.Unlock()

	for _, h := range hubs.m {
		h.broadcast <- input
	}

	return
}

//Multicast sends a message to all hubs listed in []ids.
//If no such hub exists, nothing happens.
func Multicast(message []byte, ids []string) {
	hubs.mu.RLock()
	defer hubs.mu.RUnlock()
	fmt.Printf("%+v\n", ids)
	for _, id := range ids {
		if hubs.m[id] != nil {
			hubs.m[id].Broadcast(message)
		}
	}
}

//GetHub retrieves the hub with a given ID from the hubMap.
//If no such hub exists, it creates it.
func GetHub(id string) *hub {
	hubs.mu.RLock()

	//Hub has already been created
	if hubs.m[id] != nil {
		defer hubs.mu.RUnlock()
		return hubs.m[id]
	}
	hubs.mu.RUnlock()

	//Hub has not been created
	hubs.mu.Lock()
	defer hubs.mu.Unlock()

	h := &hub{
		id:          id,
		broadcast:   make(chan []byte, cfg.broadcastMessageQueueSize), //Guarantee up to 256 messages in order
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: connectionMap{m: make(map[*connection]struct{})},
	}

	hubs.m[id] = h
	go h.run()

	return h
}

//connectionMap holds a list of connections attached to a hub
type connectionMap struct {
	m  map[*connection]struct{}
	mu sync.RWMutex
}

//hub contains all information needed to maintain a hub of communicating connections
type hub struct {
	id string

	// Registered connections.
	connections connectionMap

	// Inbound messages from the connections.
	//The buffer, if any, guarantees the number of
	//messages which will be received by every client in order
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

func (h *hub) Register(c *connection) {
	h.register <- c
}

func (h *hub) Unregister(c *connection) {
	h.unregister <- c
}

func (h *hub) Broadcast(s []byte) {
	h.broadcast <- s
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

			//If not a goroutine messages will be received by each client in order
			//(unless 1: there is a goroutine internally, or 2: hub.broadcast is unbuffered or is over its buffer)
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
	go func() {
		p, _ := Packetize("new_connection", fmt.Sprintf("%d clients currently connected to hub %s\n", numCons, h.id))
		h.broadcast <- p
		p, _ = Packetize("num_connections", numCons)
		h.broadcast <- p
	}()
}

func (h *hub) disconnect(connection *connection) {
	//could wrap these in goroutines with semaphores to make sure
	//that hub.disconnect() doesn't return until both goroutines are
	//done
	h.connections.mu.Lock()
	delete(h.connections.m, connection)
	numCons := len(h.connections.m)
	h.connections.mu.Unlock()

	connection.mu.Lock()
	connection.dead = true
	close(connection.send)
	connection.ws.Close()
	connection.mu.Unlock()

	//Unless register and unregister have a buffer, make sure any messaging during these
	//processes is concurrent.
	if numCons > 0 {
		go func() {
			p, _ := Packetize("lost_connection", fmt.Sprintf("%d clients currently connected to hub %s\n", numCons, h.id))
			h.broadcast <- p
			p, _ = Packetize("num_connections", numCons)
			h.broadcast <- p
		}()
	} else {
		defer func() {
			hubs.mu.Lock()
			defer func() { hubs.mu.Unlock() }()
			delete(hubs.m, h.id)
		}()
	}
}

func (h *hub) bcast(message []byte) {
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
		//Do not wait for one client's send before launching the next
		go conn.Send(message, finChan, h)
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
}
