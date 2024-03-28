package data

import (
	"database/sql"
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"slices"
	"sync"
	"time"
)

type GenericEvent struct {
	Type *GameEventType `json:"type"`
	Data map[string]any `json:"data"`
}
type GameEventType int

const (
	score GameEventType = iota
	foul
)

var ErrEventParseFailed = errors.New("could not parse game event")

func (e *GenericEvent) parseEvent() (GameEvent, error) {
	if e.Type == nil {
		return GameScoreEvent{}, ErrEventParseFailed
	}

	switch *e.Type {
	case score:
		event := GameScoreEvent{}
		err := checkAndAssertFromMap(e.Data, "player_pin", &event.PlayerPin)
		if err != nil {
			return GameScoreEvent{}, ErrEventParseFailed
		}
		err = checkAndAssertFromMap(e.Data, "action", &event.Action)
		if err != nil {
			return GameScoreEvent{}, ErrEventParseFailed
		}
		err = event.validate()
		if err != nil {
			return GameScoreEvent{}, ErrEventParseFailed
		}
		return event, nil
	case foul:
		event := GameFoulEvent{}
		err := checkAndAssertFromMap(e.Data, "player_pin", &event.PlayerPin)
		if err != nil {
			return GameFoulEvent{}, ErrEventParseFailed
		}
		err = checkAndAssertFromMap(e.Data, "type", &event.Type)
		if err != nil {
			return GameFoulEvent{}, ErrEventParseFailed
		}
		err = checkAndAssertFromMap(e.Data, "is_team_foul", &event.IsTeamFoul)
		if err != nil {
			return GameFoulEvent{}, ErrEventParseFailed
		}
		return event, nil
	}
	return GameFoulEvent{}, nil
}

type GameEvent interface {
	Execute(hub *GameHub)
}

var ErrEventValidationFailed = errors.New("event validation failed")

type GameScoreEvent struct {
	PlayerPin string
	Action    GameScoreAction
}
type GameScoreAction int

const (
	foul_shot GameScoreAction = iota
	two_pointer
	three_pointer
)

func (e GameScoreEvent) validate() error {
	if e.Action < 0 || e.Action > 2 {
		return ErrEventValidationFailed
	}
	return nil
}

func (e GameScoreEvent) Execute(h *GameHub) {
	statCol := h.GameInProgress.playerStats[e.PlayerPin]
	statCol.mu.Lock()
	switch e.Action {
	case foul_shot:
		statCol.stats["pts"] += 1
	case two_pointer:
		statCol.stats["pts"] += 2
	case three_pointer:
		statCol.stats["pts"] += 3
	}
	statCol.mu.Unlock()

	for watcher := range h.watchers {
		select {
		case watcher.Receive <- e:
		default:
			close(watcher.Receive)
			delete(h.watchers, watcher)
		}
	}
}

type GameFoulEvent struct {
	PlayerPin  string
	IsTeamFoul bool
	Type       GameFoulType
}

type GameFoulType int

const (
	common GameFoulType = iota
	flagrant
	loose_ball
)

func (e GameFoulEvent) Execute(h *GameHub) {
	return
}

type GameInProgress struct {
	playerStats map[string]*PlayerStats
}

type PlayerStats struct {
	stats map[string]int
	mu    sync.Mutex
}

type GameHub struct {
	AllowedKeepers []int64
	GameInProgress *GameInProgress
	keepers        map[*Keeper]bool
	watchers       map[*Watcher]bool
	Events         chan GameEvent
	JoinWatcher    chan *Watcher
	JoinKeeper     chan *Keeper
	LeaveWatcher   chan *Watcher
	LeaveKeeper    chan *Keeper
}

func NewGameHub(g *Game) *GameHub {
	allowedKeepers := []int64{g.UserID}

	statsMap := make(map[string]*PlayerStats)
	gamePlayers := append(g.Teams.Home.Players, g.Teams.Away.Players...)
	for _, p := range gamePlayers {
		statsMap[p.PinId.Pin] = &PlayerStats{
			stats: map[string]int{
				"pts": 0,
				"reb": 0,
				"ast": 0,
			},
			mu: sync.Mutex{},
		}
	}

	return &GameHub{
		AllowedKeepers: allowedKeepers,
		GameInProgress: &GameInProgress{playerStats: statsMap},
		keepers:        make(map[*Keeper]bool),
		watchers:       make(map[*Watcher]bool),
		Events:         make(chan GameEvent),
		JoinWatcher:    make(chan *Watcher),
		JoinKeeper:     make(chan *Keeper),
		LeaveKeeper:    make(chan *Keeper),
		LeaveWatcher:   make(chan *Watcher),
	}
}

func (h *GameHub) Run() {
	for {
		select {
		case watcher := <-h.JoinWatcher:
			h.watchers[watcher] = true
		case watcher := <-h.LeaveWatcher:
			if _, ok := h.watchers[watcher]; ok {
				delete(h.watchers, watcher)
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
			event.Execute(h)
		}
	}
}

func (h *GameHub) handleEvent(e *GameEvent) (map[string]int, error) {
	return nil, nil
}

type Keeper struct {
	Hub    *GameHub
	Conn   *websocket.Conn
	UserID int64
}

// TODO return close error on game hub and close connections and goroutines when closed

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
		err := k.Conn.ReadJSON(&genericEvent)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		event, err := genericEvent.parseEvent()
		if err != nil {
			return
		}
		k.Hub.Events <- event
	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

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
	Receive chan GameEvent
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
		case event, ok := <-w.Receive:
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
			jsonEvent, err := json2.Marshal(event)
			if err != nil {
				return
			}
			writer.Write(jsonEvent)

			// Add queued chat messages to the current websocket message.
			n := len(w.Receive)
			for i := 0; i < n; i++ {
				writer.Write(newline)
				jsonEvent, err := json2.Marshal(<-w.Receive)
				if err != nil {
					return
				}
				writer.Write(jsonEvent)
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

func (m *GameModel) Start(userID int64, gamePin string, gamesInProgress map[string]*GameHub) (
	*Game, error) {
	game, err := m.Get(userID, gamePin)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		}
	}

	hub := NewGameHub(game)
	go hub.Run()
	gamesInProgress[gamePin] = hub

	return game, nil
}

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
