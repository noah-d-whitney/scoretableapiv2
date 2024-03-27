package data

import (
	"database/sql"
	"errors"
	"github.com/gorilla/websocket"
)

type GameEvent struct {
	PlayerPin string `json:"player_pin"`
	Stat      string `json:"stat"`
	Action    string `json:"action"`
}

type GameHub struct {
	UserID   int64
	keepers  map[*Keeper]bool
	watchers map[*Watcher]bool
	Events   chan []byte
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
