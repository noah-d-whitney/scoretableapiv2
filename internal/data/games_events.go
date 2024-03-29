package data

import (
	json2 "encoding/json"
	"errors"
	"fmt"
)

var (
	ErrEventParseFailed      = errors.New("could not parse game event")
	ErrEventValidationFailed = errors.New("event validation failed")
)

type GameEvent interface {
	execute(hub *GameHub)
}

type GameEventType int

const (
	score GameEventType = iota
	foul
)

type GenericEvent map[string]any

func (e GenericEvent) parseEvent() (GameEvent, error) {
	eventType, err := checkAndAssertIntFromMap(e, "type")
	if err != nil {
		fmt.Printf("%s", err)
		return GameScoreEvent{}, err
	}
	fmt.Printf("event type: %d", eventType)

	switch GameEventType(eventType) {
	case score:
		event := &GameScoreEvent{}
		pin, err := checkAndAssertStringFromMap(e, "player_pin")
		if err != nil {
			return GameScoreEvent{}, ErrEventParseFailed
		}
		event.PlayerPin = pin

		action, err := checkAndAssertIntFromMap(e, "action")
		if err != nil {
			return GameScoreEvent{}, ErrEventParseFailed
		}
		event.Action = GameScoreAction(action)

		err = event.validate()
		if err != nil {
			return GameScoreEvent{}, ErrEventParseFailed
		}
		return event, nil
	case foul:
		event := GameFoulEvent{}
		pin, err := checkAndAssertStringFromMap(e, "player_pin")
		if err != nil {
			return GameFoulEvent{}, ErrEventParseFailed
		}
		event.PlayerPin = pin

		foulType, err := checkAndAssertIntFromMap(e, "foul_type")
		if err != nil {
			return GameFoulEvent{}, ErrEventParseFailed
		}
		event.FoulType = GameFoulType(foulType)

		isTeamFoul, err := checkAndAssertBoolFromMap(e, "is_team_foul")
		if err != nil {
			return GameFoulEvent{}, ErrEventParseFailed
		}
		event.IsTeamFoul = isTeamFoul
		return event, nil
	}
	return GameFoulEvent{}, nil
}

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

func (e GameScoreEvent) generateClientMessage(h *GameHub) ([]byte, error) {
	bytes, err := json2.Marshal(h.GameInProgress.playerStats[e.PlayerPin])
	if err != nil {
		return nil, err
	}
	fmt.Printf("\nOUTPUT: %s\n", string(bytes))

	return bytes, nil
}

func (e GameScoreEvent) execute(h *GameHub) {
	statCol := h.GameInProgress.playerStats[e.PlayerPin]
	statCol.mu.Lock()
	switch e.Action {
	case foul_shot:
		statCol.Ftm += 1
	case two_pointer:
		statCol.TwoPointers += 1
	case three_pointer:
		statCol.ThreePointers += 1
	}
	statCol.mu.Unlock()

	message, err := e.generateClientMessage(h)
	if err != nil {
		return
	}

	for watcher := range h.watchers {
		select {
		case watcher.Receive <- message:
		default:
			close(watcher.Receive)
			delete(h.watchers, watcher)
		}
	}
}

type GameFoulEvent struct {
	PlayerPin  string
	IsTeamFoul bool
	FoulType   GameFoulType
}

type GameFoulType int

const (
	common GameFoulType = iota
	flagrant
	loose_ball
)

func (e GameFoulEvent) execute(h *GameHub) {
	return
}
