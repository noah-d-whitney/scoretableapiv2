package gamehub

import (
	"ScoreTableApi/internal/clock"
	"ScoreTableApi/internal/stats"
	json2 "encoding/json"
	"errors"
	"fmt"
)

type GameEvent interface {
	execute(hub *Hub)
}

type GameEventType int

const (
	stat GameEventType = iota
	gameClock
)

type GenericEvent map[string]any

func (e GenericEvent) parseEvent() (GameEvent, error) {
	eventType, err := checkAndAssertIntFromMap(e, "type")
	if err != nil {
		fmt.Printf("%s", err)
		return GameStatEvent{}, err
	}
	fmt.Printf("event type: %d", eventType)

	switch GameEventType(eventType) {
	case stat:
		event := &GameStatEvent{}
		pin, err := checkAndAssertStringFromMap(e, "player_pin")
		if err != nil {
			return GameStatEvent{}, ErrEventParseFailed
		}
		event.PlayerPin = pin

		action, err := checkAndAssertIntFromMap(e, "action")
		if err != nil {
			return GameStatEvent{}, ErrEventParseFailed

		}
		event.Action = GameStatAction(action)
		selectedStat, err := checkAndAssertStringFromMap(e, "stat")
		if err != nil {
			return GameStatEvent{}, ErrEventParseFailed
		}
		event.Stat = stats.PrimitiveStat(selectedStat)

		err = event.validate()
		if err != nil {
			return GameStatEvent{}, ErrEventParseFailed
		}
		return event, nil
	case gameClock:
		event := &GameClockEvent{}

		action, err := checkAndAssertIntFromMap(e, "action")
		if err != nil {
			return GameClockEvent{}, ErrEventParseFailed
		}
		event.Action = clock.EventType(action)

		value, _ := checkAndAssertStringFromMap(e, "value")
		if value == "" {
			event.Value = nil
		} else {
			event.Value = &value
		}

		err = event.validate()
		if err != nil {
			return GameClockEvent{}, ErrEventParseFailed
		}
		return event, nil
	}

	return GameStatEvent{}, nil
}

// TODO make anonymous event to return that executes and sends problems

type GameStatEvent struct {
	PlayerPin string
	Stat      stats.PrimitiveStat
	Action    GameStatAction
}

type GameStatAction int

const (
	add GameStatAction = iota
	subtract
)

func (e GameStatEvent) validate() error {
	if e.Action < 0 || e.Action > 1 {
		return ErrEventValidationFailed
	}
	return nil
}

func (e GameStatEvent) generateClientMessage(h *Hub) ([]byte, error) {
	bytes, err := json2.Marshal(h.Stats.GetDto())
	if err != nil {
		return nil, err
	}
	fmt.Printf("\nOUTPUT: %s\n", string(bytes))

	return bytes, nil
}

func (e GameStatEvent) execute(h *Hub) {
	switch e.Action {
	case add:
		h.Stats.Add(e.PlayerPin, e.Stat, 1)
	case subtract:
		h.Stats.Add(e.PlayerPin, e.Stat, -1)
	}

	message, err := e.generateClientMessage(h)
	if err != nil {
		return
	}

	h.ToAllWatchers(message)
}

type GameClockEvent struct {
	Action clock.EventType
	Value  *string
}

func (e GameClockEvent) validate() error {
	switch e.Action {
	case clock.Play, clock.Pause, clock.Reset:
		if e.Value != nil {
			return errors.Join(ErrEventValidationFailed,
				errors.New("clock event with specified action must have nil Value field"))
		}
		return nil
	case clock.PeriodChange:
		if e.Value == nil {
			return errors.Join(ErrEventValidationFailed,
				errors.New("clock event with specified action cannot have null Value field"))
		}
		if *e.Value == "+" || *e.Value == "-" {
			return nil
		} else {
			return errors.Join(ErrEventValidationFailed,
				errors.New("clock event with specified action cannot have specified Value field"))
		}
	case clock.Set:
		if e.Value == nil {
			return errors.Join(ErrEventValidationFailed,
				errors.New("clock event with specified action cannot have null Value field"))
		}
		return nil
	default:
		return ErrEventValidationFailed
	}
}

func (e GameClockEvent) execute(h *Hub) {
	switch clock.EventType(e.Action) {
	case clock.Play:
		h.Clock.Play()
	case clock.Pause:
		h.Clock.Pause()
	case clock.Reset:
		h.Clock.Reset()
	case clock.Set:
		h.Clock.Set(clock.ClockDuration(*e.Value))
	case clock.PeriodChange:
		switch *e.Value {
		case "+":
			h.Clock.ChangePeriod(1)
		case "-":
			h.Clock.ChangePeriod(-1)
		}
	default:
		return
	}
}