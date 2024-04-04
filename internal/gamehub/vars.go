package gamehub

import (
	"errors"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 1 * time.Minute

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline                  = []byte{'\n'}
	space                    = []byte{' '}
	ErrEventParseFailed      = errors.New("could not parse game event")
	ErrEventValidationFailed = errors.New("event validation failed")
)
