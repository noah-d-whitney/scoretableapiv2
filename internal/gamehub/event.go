package gamehub

import (
	"ScoreTableApi/internal/clock"
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/stats"
	json2 "encoding/json"
	"fmt"
)

type GameEvent interface {
	execute(hub *Hub)
}

type GameEventType int

const (
	stat GameEventType = iota
	gameClock
	substitution
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
		event.Action = clock.Control(action)

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
	case substitution:
		event := &GameSubstitutionEvent{}

		side, err := checkAndAssertIntFromMap(e, "side")
		if err != nil {
			return GameSubstitutionEvent{}, ErrEventParseFailed
		}
		event.Side = data.GameTeamSide(side)

		out, err := checkAndAssertStringFromMap(e, "out")
		if err != nil {
			return GameSubstitutionEvent{}, ErrEventParseFailed
		}
		event.Out = out

		in, err := checkAndAssertStringFromMap(e, "in")
		if err != nil {
			return GameSubstitutionEvent{}, ErrEventParseFailed
		}
		event.In = in

		err = event.validate()
		if err != nil {
			return GameSubstitutionEvent{}, ErrEventParseFailed
		}

		return event, nil
	}

	return GameStatEvent{}, nil
}

// TODO: make anonymous event to return that executes and sends problems

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
	if !h.Lineups.isActive(e.PlayerPin) {
		return
	}
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
	Action clock.Control
	Value  *string
}

func (e GameClockEvent) validate() error {
	return nil
}

func (e GameClockEvent) execute(h *Hub) {
	h.Clock.Controller <- e.Action
}

type GameSubstitutionEvent struct {
	Side data.GameTeamSide
	In   string
	Out  string
}

func (e GameSubstitutionEvent) validate() error {
	if e.In == "" || e.Out == "" {
		return ErrEventValidationFailed
	}
	return nil
}

func (e GameSubstitutionEvent) execute(h *Hub) {
	if h.Clock.GetState() == clock.StatePlaying {
		return
	}

	h.Lineups.substitution(e.Side, e.Out, e.In)
	msg := h.toByteArr(envelope{
		"active": h.Lineups.getActive(),
		"bench":  h.Lineups.getBench(),
		"Dnp":    h.Lineups.getDnp(),
		"subs": map[string]string{
			"out": e.Out,
			"in":  e.In,
		},
	})
	h.ToAllKeepers(msg)
}
