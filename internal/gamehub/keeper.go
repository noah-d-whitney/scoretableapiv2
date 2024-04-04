package gamehub

import (
	json2 "encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Keeper struct {
	Hub     *Hub
	Conn    *websocket.Conn
	UserID  int64
	Receive chan []byte
	Close   chan bool
}

func NewKeeper(userID int64, hub *Hub, conn *websocket.Conn) *Keeper {
	return &Keeper{
		Hub:     hub,
		Conn:    conn,
		UserID:  userID,
		Receive: make(chan []byte),
		Close:   make(chan bool),
	}
}

// TODO return close error on game hub and close connections and goroutines when closed
func (k *Keeper) WriteEvents() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		k.Hub.LeaveKeeper(k.UserID)
	}()
	for {
		select {
		case <-ticker.C:
			k.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := k.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-k.Close:
			closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
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
		k.Hub.LeaveKeeper(k.UserID)
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
