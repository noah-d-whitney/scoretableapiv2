package data

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
)

type GameInProgress struct {
	playerStats map[string]*PlayerStats
}

type PlayerStats struct {
	Points        *PlayerGetterStat `json:"pts"`
	ThreePointers int               `json:"3ptm"`
	TwoPointers   int               `json:"2pta"`
	Ftm           int               `json:"ftm"`
	Stats         map[string]int    `json:"-"`
	mu            sync.Mutex
}

type PlayerGetterStat struct {
	statSrc *PlayerStats
	getFunc func(statSrc *PlayerStats) int
}

func (pgs PlayerGetterStat) get() int {
	return pgs.getFunc(pgs.statSrc)
}

func (pgs PlayerGetterStat) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(pgs.get())), nil
}

func newPointsGetter(src *PlayerStats) *PlayerGetterStat {
	return &PlayerGetterStat{
		statSrc: src,
		getFunc: func(src *PlayerStats) int {
			var points int
			points += src.Ftm
			points += src.TwoPointers * 2
			points += src.ThreePointers * 3
			return points
		},
	}
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
		playerStats := &PlayerStats{
			ThreePointers: 0,
			TwoPointers:   0,
			Ftm:           0,
			Stats:         nil,
			mu:            sync.Mutex{},
		}
		playerStats.Points = newPointsGetter(playerStats)
		statsMap[p.PinId.Pin] = playerStats
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
			event.execute(h)
		case err := <-h.Errors:
			fmt.Printf("\nHUB ERROR: %s\n", err.Error())
			for k := range h.keepers {
				k.Close <- err
			}
			for w := range h.watchers {
				w.Close <- err
			}
		}
	}
}
