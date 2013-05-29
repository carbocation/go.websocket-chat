package wshub

import (
	"time"
)

type config struct {
	writeWait                 time.Duration
	readWait                  time.Duration
	pingPeriod                time.Duration
	maxMessageSize            int64
	broadcastMessageQueueSize int64
}

func Initialize(writeWait, readWait, pingPeriod time.Duration, maxMessageSize, broadcastMessageSize int64) {
	cfg = &config{writeWait: writeWait,
		readWait:                  readWait,
		pingPeriod:                pingPeriod,
		maxMessageSize:            maxMessageSize,
		broadcastMessageQueueSize: broadcastMessageSize,
	}
}
