package gamehub

import (
	"ScoreTableApi/internal/clock"
	"ScoreTableApi/internal/stats"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"slices"
)

var (
	ErrKeeperNotAuthorized = errors.New("keeper not authorized")
)

type Hub struct {
	AllowedKeepers []int64
	Stats          *stats.GameStatline
	Clock          *clock.GameClock
	keepers        map[*keeper]bool
	Watchers       map[*Watcher]bool
	Events         chan GameEvent
	Errors         chan error
	JoinWatcher    chan *Watcher
	joinKeeper     chan *keeper
	LeaveWatcher   chan *Watcher
	LeaveKeeper    chan *keeper
}

func (h *Hub) JoinKeeper(userID int64, conn *websocket.Conn) error {
	keeper := newKeeper(userID, h, conn)
	if !slices.Contains(h.AllowedKeepers, keeper.UserID) {
		return ErrKeeperNotAuthorized
	}

	h.joinKeeper <- keeper
	go keeper.ReadEvents()
	go keeper.WriteEvents()

	return nil
}

// TODO pass in blueprint on create statline, send out list of possible stats to keeper and client

func (h *Hub) Run() {
	for {
		select {
		case watcher := <-h.JoinWatcher:
			h.Watchers[watcher] = true
		case watcher := <-h.LeaveWatcher:
			if _, ok := h.Watchers[watcher]; ok {
				delete(h.Watchers, watcher)
				close(watcher.Receive)
			}
		case keeper := <-h.JoinKeeper:
			if slices.Contains(h.AllowedKeepers, keeper.UserID) {
				h.keepers[keeper] = true
			}
		case keeper := <-h.LeaveKeeper:
			if _, ok := h.keepers[keeper]; ok {
				delete(h.keepers, keeper)
			}
		case event := <-h.Events:
			fmt.Printf("event from hub: %v", event)
			event.execute(h)
		case err := <-h.Errors:
			fmt.Printf("\nHUB ERROR: %s\n", err.Error())
			for k := range h.keepers {
				k.Close <- err
			}
			for w := range h.Watchers {
				w.Close <- err
			}
		}
	}
}

func (h *Hub) ToAllWatchers(msg []byte) {
	for watcher := range h.Watchers {
		select {
		case watcher.Receive <- msg:
		default:
			close(watcher.Receive)
			delete(h.Watchers, watcher)
		}
	}
}

func (h *Hub) PipeClockToWatchers() {
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
