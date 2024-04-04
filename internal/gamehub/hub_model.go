package gamehub

import (
	"ScoreTableApi/internal/clock"
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/stats"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var (
	ErrGameNotFound = errors.New("game not found")
	ErrTwoTeams     = errors.New("game must have two teams to start")
	upgrader        = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type HubModel struct {
	active map[string]*Hub
	model  *data.GameModel
}

func NewModel(model *data.GameModel) HubModel {
	return HubModel{
		active: make(map[string]*Hub),
		model:  model,
	}
}

func (m *HubModel) StartGame(g *data.Game) (*Hub, error) {
	err := m.validateGame(g)
	if err != nil {
		return nil, err
	}

	err = m.model.StartGameInDB(g)
	if err != nil {
		return nil, err
	}

	hub := &Hub{
		AllowedKeepers: []int64{g.UserID},
		Game:           g,
		Stats:          stats.NewGameStatline(g.HomePlayerPins, g.AwayPlayerPins, stats.Simple),
		keepers:        make(map[int64]*Keeper),
		Watchers:       make(map[*Watcher]bool),
		Events:         make(chan GameEvent),
		Errors:         make(chan error),
	}

	var c *clock.GameClock
	if g.Type == "timed" {
		c = clock.NewGameClock(clock.Config{
			PeriodLength: time.Duration(*g.PeriodLength),
			PeriodCount:  *g.PeriodCount,
			OtDuration:   time.Duration(*g.PeriodLength) / 2,
		})
	} else {
		c = clock.NewGameClock(clock.Config{
			PeriodLength: 0,
			PeriodCount:  0,
			OtDuration:   0,
		})
	}
	hub.Clock = c

	m.active[g.PinID.Pin] = hub
	go hub.Run()

	return hub, nil
}

func (m *HubModel) WatcherJoinGame(pin string, wr http.ResponseWriter, r *http.Request) (*Watcher,
	error) {
	if _, ok := m.active[pin]; !ok {
		return nil, ErrGameNotFound
	}
	h := m.active[pin]

	conn, err := upgrader.Upgrade(wr, r, nil)
	if err != nil {
		return nil, err
	}

	w := h.JoinWatcher(conn)
	return w, nil
}

// TODO validate game before starting

func (m *HubModel) validateGame(game *data.Game) error {
	if game.HomeTeamPin == nil || game.AwayTeamPin == nil {
		return ErrTwoTeams
	}
	return nil
}
