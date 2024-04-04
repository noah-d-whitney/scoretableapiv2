package data

import (
	"database/sql"
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Keeper struct {
	Hub    *GameHub
	Conn   *websocket.Conn
	UserID int64
	Close  chan error
}

// TODO return close error on game hub and close connections and goroutines when closed
func (k *Keeper) WriteEvents() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		k.Hub.LeaveKeeper <- k
		k.Conn.Close()
	}()
	for {
		select {
		case <-ticker.C:
			k.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := k.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case closeErr := <-k.Close:
			closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, closeErr.Error())
			writer, err := k.Conn.NextWriter(websocket.CloseMessage)
			if err != nil {
				return
			}
			writer.Write(closeMessage)
			writer.Close()
			return
		}
	}
}

func (k *Keeper) ReadEvents() {
	defer func() {
		k.Hub.LeaveKeeper <- k
		k.Conn.Close()
	}()
	for {
		k.Conn.SetReadLimit(maxMessageSize)
		k.Conn.SetReadDeadline(time.Now().Add(pongWait))
		k.Conn.SetPongHandler(func(string) error {
			k.Conn.SetReadDeadline(time.Now().Add(
				pongWait))
			return nil
		})
		var genericEvent GenericEvent
		_, bytes, err := k.Conn.ReadMessage()
		if err != nil {
			k.Hub.Errors <- err
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		err = json2.Unmarshal(bytes, &genericEvent)
		if err != nil {
			k.Hub.Errors <- err
		}
		event, err := genericEvent.parseEvent()
		if err != nil {
			k.Hub.Errors <- err
		}
		k.Hub.Events <- event
	}
}

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
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Watcher struct {
	Hub     *GameHub
	Conn    *websocket.Conn
	Receive chan []byte
	Close   chan error
}

func (w *Watcher) WriteEvents() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		w.Hub.LeaveWatcher <- w
		w.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-w.Receive:

			fmt.Printf("event from writeEvent: %v", message)
			w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := w.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			writer.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(w.Receive)
			for i := 0; i < n; i++ {
				writer.Write(newline)
				writer.Write(<-w.Receive)
			}

			if err := writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := w.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case closeErr := <-w.Close:
			closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, closeErr.Error())
			writer, err := w.Conn.NextWriter(websocket.CloseMessage)
			if err != nil {
				return
			}
			writer.Write(closeMessage)
			writer.Close()
			return
		}
	}
}

func (h *GameHub) pipeClockToWatchers() {
	for {
		select {
		case e := <-h.Clock.C:
			msg := []byte(e.Value)
			h.ToAllWatchers(msg)
		case <-h.Errors:
			return
		}
	}
}

func (m *GameModel) Start(userID int64, gamePin string, gamesInProgress map[string]*GameHub) (
	*GameHub, error) {
	game, err := m.Get(userID, gamePin)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		}
	}

	hub := NewGameHub(game)
	go hub.Run()
	if game.Type == GameTypeTimed {
		go hub.pipeClockToWatchers()
	}
	gamesInProgress[gamePin] = hub

	return hub, nil
}

// TODO fix game cleanup

func (m *GameModel) End(gamePin string, gamesInProgress map[string]*GameHub) {
	game, ok := gamesInProgress[gamePin]
	if !ok {
		return
	}
	for watcher, _ := range game.watchers {
		watcher.Close <- fmt.Errorf("connection closed")
	}
	delete(gamesInProgress, gamePin)
}
