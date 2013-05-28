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

//Package-global config lives here:
var cfg *config

func init() {
	Initialize(10*time.Second,
		60*time.Second,
		60*time.Second*9/10,
		4096,
		256)
}
