package gamehub

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type Watcher struct {
	Hub     *Hub
	Conn    *websocket.Conn
	Receive chan []byte
	Error   chan error
}

func newWatcher(hub *Hub, conn *websocket.Conn) *Watcher {
	return &Watcher{
		Hub:     hub,
		Conn:    conn,
		Receive: make(chan []byte),
		Error:   make(chan error),
	}
}

func (w *Watcher) WriteEvents() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		w.Hub.LeaveWatcher(w)
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
		case closeErr := <-w.Error:
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
