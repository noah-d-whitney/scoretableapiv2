package pins

import (
	"errors"
	"math/rand/v2"
	"strconv"
)

var (
	ErrDuplicatePin = errors.New("duplicate pin")
	letterRunes     = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	PinScopeTeams   = "teams"
	PinScopePlayers = "players"
	PinScopeGames   = "games"
	PinScopeLeagues = "leagues"
)

type Pin struct {
	ID    int64
	Pin   string
	Scope string
}

func (p Pin) MarshalJSON() ([]byte, error) {
	jsonValue := strconv.Quote(p.Pin)
	return []byte(jsonValue), nil
}

func GeneratePin(l int) string {
	b := make([]rune, l)
	for i := range b {
		b[i] = letterRunes[rand.IntN(len(letterRunes))]
	}
	return string(b)
}
