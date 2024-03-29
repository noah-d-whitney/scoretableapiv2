package data

import (
	"fmt"
	"slices"
	"sync"
)

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
	Errors         chan error
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
		Errors:         make(chan error),
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
			fmt.Printf("event from hub: %v", event)
			event.Execute(h)
		}
	}
}

func (h *GameHub) handleEvent(e *GameEvent) (map[string]int, error) {
	return nil, nil
}
