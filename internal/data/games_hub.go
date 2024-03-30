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
	Pts     PlayerGetterStat `json:"pts"`
	FgPer   PlayerGetterStat `json:"fg%"`
	ThPtPer PlayerGetterStat `json:"3pt%"`
	FtPer   PlayerGetterStat `json:"ft%"`
	ThPtA   int              `json:"3pta"`
	ThPtM   int              `json:"3ptm"`
	TwPtA   int              `json:"2pta"`
	TwPtM   int              `json:"2ptm"`
	FtA     int              `json:"fta"`
	FtM     int              `json:"ftm"`
	mu      sync.Mutex
}

func NewPlayerStats() *PlayerStats {
	stats := &PlayerStats{
		FgPer:   nil,
		ThPtPer: nil,
		FtPer:   nil,
		ThPtA:   0,
		ThPtM:   0,
		TwPtA:   0,
		TwPtM:   0,
		FtA:     0,
		FtM:     0,
		mu:      sync.Mutex{},
	}

	stats.Pts = newPointsGetter(stats)
}

type PlayerGetterStat interface {
	get() interface{}
	MarshalJSON() ([]byte, error)
}

type PlayerPointsGetter struct {
	src *PlayerStats
}

func (ppg *PlayerPointsGetter) get() any {
	var points int
	points += ppg.src.FtM
	points += ppg.src.TwPtM * 2
	points += ppg.src.ThPtM * 3
	return points
}

func (ppg *PlayerPointsGetter) MarshalJSON() ([]byte, error) {
	value := ppg.get().(int)
	return []byte(strconv.Itoa(value)), nil
}

func newPointsGetter(src *PlayerStats) *PlayerPointsGetter {
	return &PlayerPointsGetter{
		src: src,
	}
}

//func newThPtPerGetter(src *PlayerStats) *PlayerGetterStat {
//	return &PlayerGetterStat{
//		statSrc: src,
//		getFunc: func(src *PlayerStats) any {
//			percent := float64(src.ThPtA)/float64(src.ThPtM)
//			return percent
//		},
//	}
//}

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
		playerStats.Pts = newPointsGetter(playerStats)
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
