/*
Websocket-Hub is heavily derived from Gary Burd's example chat service,
and much of this is copyright Gary Burd, used under the Apache License,
Version 2.0 ( https://github.com/garyburd/go-websocket#license )

The remainder is copyright 2013 James Pirruccello. The entire work is
offered under the Apache License, Version 2.0.
*/
package main

import (
	"time"
)

//Hub of hubs lives here
var hubs *hubMap = &hubMap{m: map[string]*hub{}}

//Connection constants
const (
	// Time allowed to write a message to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next message from the client.
	readWait = 60 * time.Second

	// Send pings to client with this period. Must be less than readWait.
	pingPeriod = (readWait * 9) / 10

	// Maximum message size (in characters) allowed from client. They will be disconnected if larger.
	maxMessageSize = 4096

	// Maximum number of messages pending in each hub's broadcast queue
	broadcastMessageQueueSize = 256
)
