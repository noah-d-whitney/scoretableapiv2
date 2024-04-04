package gamehub

import (
	"ScoreTableApi/internal/clock"
	"ScoreTableApi/internal/stats"
	"context"
	"database/sql"
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
	db     *sql.DB
}

func NewModel(db *sql.DB) HubModel {
	return HubModel{
		active: make(map[string]*Hub),
		db:     db,
	}
}

func (m *HubModel) StartGame(pin string, userID int64) (*Hub, error) {
	g, err := m.getGame(pin, userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGameNotFound
		default:
			return nil, err
		}
	}
	err = m.validateGame(g)
	if err != nil {
		return nil, err
	}

	hub := &Hub{
		AllowedKeepers: g.allowedKeepers,
		Stats:          stats.NewGameStatline(g.homePlayerPins, g.awayPlayerPins, g.blueprint),
		keepers:        make(map[int64]*Keeper),
		Watchers:       make(map[*Watcher]bool),
		Events:         make(chan GameEvent),
		Errors:         make(chan error),
	}

	var c *clock.GameClock
	if g.gameType == "timed" {
		c = clock.NewGameClock(clock.Config{
			PeriodLength: *g.periodLength,
			PeriodCount:  *g.periodCount,
			OtDuration:   *g.periodLength / 2,
		})
	} else {
		c = nil
	}
	hub.Clock = c

	m.active[g.pin] = hub
	if hub.Clock != nil {
		go hub.PipeClockToWatchers()
	}
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

func (m *HubModel) validateGame(game *game) error {
	if game.homeTeamPin == nil || game.awayTeamPin == nil {
		return ErrTwoTeams
	}
	return nil
}

func (m *HubModel) getGame(pin string, userID int64) (*game, error) {
	stmt := `
		SELECT pin, user_id, home_team_pin, away_team_pin, home_player_pins, away_player_pins,
			period_count, period_length, score_target, type, team_size
		FROM games_view
		WHERE pin = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var g game
	err := m.db.QueryRowContext(ctx, stmt, pin, userID).Scan(
		&g.pin,
		&g.owner,
		&g.homeTeamPin,
		&g.awayTeamPin,
		&g.homePlayerPins,
		&g.awayPlayerPins,
		&g.periodCount,
		&g.periodLength,
		&g.scoreTarget,
		&g.gameType,
		&g.teamSize,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGameNotFound
		default:
			return nil, err
		}
	}

	g.blueprint = stats.Simple
	g.allowedKeepers = []int64{g.owner}

	return &g, nil
}

type game struct {
	pin            string
	owner          int64
	allowedKeepers []int64
	homeTeamPin    *string
	awayTeamPin    *string
	homePlayerPins []string
	gameType       string
	awayPlayerPins []string
	periodCount    *int64
	periodLength   *time.Duration
	scoreTarget    *int64
	teamSize       int64
	blueprint      stats.Blueprint
}
