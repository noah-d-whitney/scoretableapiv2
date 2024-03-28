package data

import (
	"database/sql"
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type GameEvent struct {
	PlayerPin string `json:"player_pin"`
	Stat      string `json:"stat"`
	Action    string `json:"action"`
}

type GameInProgress struct {
	stats map[string]*PlayerStats
}

type PlayerStats struct {
	stats map[string]int
	mu    sync.Mutex
}

type GameHub struct {
	AllowedKeepers []int64
	Stats          *GameInProgress
	keepers        map[*Keeper]bool
	watchers       map[*Watcher]bool
	Events         chan []byte
	JoinWatcher    chan *Watcher
	JoinKeeper     chan *Keeper
}

func NewGameHub(g *Game) *GameHub {
	allowedKeepers := []int64{g.UserID}

	var statsMap map[string]*PlayerStats
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
		Stats:          &GameInProgress{stats: statsMap},
		keepers:        make(map[*Keeper]bool),
		watchers:       make(map[*Watcher]bool),
		Events:         make(chan []byte),
		JoinWatcher:    make(chan *Watcher),
		JoinKeeper:     make(chan *Keeper),
	}
}

func (h *GameHub) Run() {

}

type Keeper struct {
	Hub  *GameHub
	Conn *websocket.Conn
	Send chan []byte
}

type Watcher struct {
	Hub     *GameHub
	Conn    *websocket.Conn
	Receive chan []byte
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

	gamesInProgress[gamePin] = &GameHub{
		keepers:  make(map[*Keeper]bool),
		watchers: make(map[*Watcher]bool),
		Events:   make(chan []byte),
	}

	return game, nil
}
