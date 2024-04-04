package data

import (
	"ScoreTableApi/internal/clock"
	"ScoreTableApi/internal/stats"
	"fmt"
	"slices"
)

type GameHub struct {
	AllowedKeepers []int64
	Stats          *stats.GameStatline
	Clock          *clock.GameClock
	keepers        map[*Keeper]bool
	watchers       map[*Watcher]bool
	Events         chan GameEvent
	Errors         chan error
	JoinWatcher    chan *Watcher
	JoinKeeper     chan *Keeper
	LeaveWatcher   chan *Watcher
	LeaveKeeper    chan *Keeper
}

// TODO pass in blueprint on create statline, send out list of possible stats to keeper and client
func NewGameHub(g *Game) *GameHub {
	allowedKeepers := []int64{g.UserID}
	homePins, awayPins := g.getPlayerPins()

	return &GameHub{
		AllowedKeepers: allowedKeepers,
		Stats:          stats.NewGameStatline(homePins, awayPins, stats.Standard),
		Clock: clock.NewGameClock(clock.Config{
			PeriodLength: g.PeriodLength.Duration(),
			PeriodCount:  *g.PeriodCount,
			OtDuration:   g.PeriodLength.Duration() / 2,
		}),
		keepers:      make(map[*Keeper]bool),
		watchers:     make(map[*Watcher]bool),
		Events:       make(chan GameEvent),
		Errors:       make(chan error),
		JoinWatcher:  make(chan *Watcher),
		JoinKeeper:   make(chan *Keeper),
		LeaveKeeper:  make(chan *Keeper),
		LeaveWatcher: make(chan *Watcher),
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

func (h *GameHub) ToAllWatchers(msg []byte) {
	for watcher := range h.watchers {
		select {
		case watcher.Receive <- msg:
		default:
			close(watcher.Receive)
			delete(h.watchers, watcher)
		}
	}
}
