package gamehub

import (
	json2 "encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type keeper struct {
	Hub     *Hub
	Conn    *websocket.Conn
	UserID  int64
	Receive chan []byte
	Close   chan error
}

func newKeeper(userID int64, hub *Hub, conn *websocket.Conn) *keeper {
	return &keeper{
		Hub:     hub,
		Conn:    conn,
		UserID:  userID,
		Receive: make(chan []byte),
		Close:   make(chan error),
	}
}

// TODO return close error on game hub and close connections and goroutines when closed
func (k *keeper) WriteEvents() {
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

func (k *keeper) ReadEvents() {
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
