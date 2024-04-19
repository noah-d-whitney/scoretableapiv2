package gamehub

import (
	"ScoreTableApi/internal/clock"
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/stats"
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"slices"
)

var (
	ErrKeeperNotAuthorized = errors.New("Keeper not authorized")
)

type Hub struct {
	AllowedKeepers []int64
	Game           *data.Game
	Stats          *stats.GameStatline
	Clock          *clock.GameClock
	//Plays          *PlayEngine
	//Model for saving stats
	Lineups  *lineupManager
	keepers  map[int64]*Keeper
	Watchers map[*Watcher]bool
	Events   chan GameEvent
	Errors   chan error
}

func (h *Hub) JoinKeeper(userID int64, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.Errors <- err
	}

	k := Keeper{
		Hub:     h,
		Conn:    conn,
		UserID:  userID,
		Receive: make(chan []byte),
		Close:   make(chan bool),
	}
	if !slices.Contains(h.AllowedKeepers, k.UserID) {
		return ErrKeeperNotAuthorized
	}

	h.keepers[k.UserID] = &k
	go k.ReadEvents()
	go k.WriteEvents()

	welcomeData := h.toByteArr(envelope{
		"stats":    h.Stats.GetDto(),
		"clock":    h.Clock.Get(),
		"period":   h.Clock.GetPeriod(),
		"game":     h.Game,
		"timeouts": h.Clock.GetTimeouts(),
		"active":   h.Lineups.getActive(),
		"bench":    h.Lineups.getBench(),
		"dnp":      h.Lineups.getDnp(),
	})

	k.Receive <- welcomeData

	return nil
}

func (h *Hub) LeaveKeeper(userID int64) {
	if k, ok := h.keepers[userID]; ok {
		delete(h.keepers, userID)
		close(k.Receive)
		close(k.Close)
		k.Conn.Close()
	}
}

// TODO make JoinWatcher receive w and r instead of conn

func (h *Hub) JoinWatcher(conn *websocket.Conn) *Watcher {
	w := newWatcher(h, conn)
	h.Watchers[w] = true
	go w.WriteEvents()

	welcomeData := h.toByteArr(envelope{
		"stats":  h.Stats.GetDto(),
		"clock":  h.Clock.Get(),
		"period": h.Clock.GetPeriod(),
		"game":   h.Game,
	})
	w.Receive <- welcomeData
	return w
}

func (h *Hub) LeaveWatcher(w *Watcher) {
	if _, ok := h.Watchers[w]; ok {
		delete(h.Watchers, w)
		close(w.Receive)
		close(w.Error)
		w.Conn.Close()
	}
}

// TODO pass in blueprint on create statline, send out list of possible stats to Keeper and client

func (h *Hub) Run() {
	for {
		select {
		case event := <-h.Events:
			fmt.Printf("event from hub: %v", event)
			event.execute(h)
		case tick := <-h.Clock.C:
			fmt.Printf("%+v\n", tick)
			msg := h.toByteArr(envelope{"clock": tick.Value})
			h.ToAllKeepers(msg)
			h.ToAllWatchers(msg)
		case err := <-h.Errors:
			fmt.Printf("\nHUB ERROR: %s\n", err.Error())
			for _, k := range h.keepers {
				k.Close <- true
			}
			for w := range h.Watchers {
				w.Error <- err
			}
		}
	}
}

func (h *Hub) ToAllWatchers(msg []byte) {
	for watcher := range h.Watchers {
		select {
		case watcher.Receive <- msg:
		default:
			h.LeaveWatcher(watcher)
		}
	}
}

func (h *Hub) ToAllKeepers(msg []byte) {
	for i, k := range h.keepers {
		select {
		case k.Receive <- msg:
		default:
			h.LeaveKeeper(i)
		}
	}
}

func (h *Hub) toByteArr(v envelope) []byte {
	bytes, _ := json2.Marshal(v)
	return bytes
}

type envelope map[string]any
