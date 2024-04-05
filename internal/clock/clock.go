package clock

import (
	"ScoreTableApi/internal/data"
	"errors"
	"fmt"
	"math"
	"strconv"
	strings2 "strings"
	"time"
)

var ErrInvalidDuration = errors.New("invalid clock duration string")

// Duration represents a string in the format "MM:SS"
type Duration string

// ToDuration converts string from format "MM:SS" to a time.Duration
func (cd Duration) ToDuration() (time.Duration, error) {
	strings := strings2.Split(string(cd), ":")
	minutes, err := strconv.Atoi(strings[0])
	if err != nil {
		return 0, errors.Join(ErrInvalidDuration, err)
	}
	seconds, err := strconv.Atoi(strings[1])
	if err != nil {
		return 0, errors.Join(ErrInvalidDuration, err)
	}
	if seconds >= 60 {
		return 0, ErrInvalidDuration
	}

	dur, err := time.ParseDuration(fmt.Sprintf("%dm%ds", minutes, seconds))
	if err != nil {
		return 0, errors.Join(ErrInvalidDuration, err)
	}

	return dur, nil
}

type State int

const (
	StateFresh State = iota
	StatePlaying
	StatePaused
	StateDone
	StateClosed
	StateTimeout
)

// GameClock keeps current game time and period. Sends string every second on C with current time
// when GameClock is running.
type GameClock struct {
	current    time.Duration
	toCurrent  time.Duration
	C          chan Event
	Controller chan Control
	state      State
	period     int64
	config     Config
	stop       chan bool
	homeTOs    int
	awayTOs    int
}

func (gc *GameClock) run() {
	for {
		select {
		case action, ok := <-gc.Controller:
			if !ok {
				return
			}
			switch action {
			case Play:
				gc.Play()
			case Pause:
				gc.Pause()
			case Reset:
				gc.Reset()
			case AddMin:
				gc.Adjust("01:00", true)
			case SubtractMin:
				gc.Adjust("01:00", false)
			case AddSec:
				gc.Adjust("00:01", true)
			case SubtractSec:
				gc.Adjust("00:01", false)
			case AddPeriod:
				gc.ChangePeriod(1)
			case SubtractPeriod:
				gc.ChangePeriod(-1)
			case CallTimeoutHome:
				gc.Timeout(data.TeamHome)
			case CallTimeoutAway:
				gc.Timeout(data.TeamAway)
			case EndTimeout:
				gc.Controller <- EndTimeout
			default:
			}
		}
	}
}

func (gc *GameClock) GetState() State {
	return gc.state
}

func (gc *GameClock) Timeout(side data.GameTeamSide) {
	switch gc.state {
	case StateClosed:
		return
	default:
		gc.stop <- true
		timeoutRoutine := func() {
			gc.state = StateTimeout

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			gc.C <- Event{
				EventType: Timeout,
				Value:     gc.Get(),
			}

			for {
				select {
				case control := <-gc.Controller:
					if control == EndTimeout {
						gc.state = StatePaused
						gc.toCurrent = gc.config.TimeoutDuration
						gc.C <- Event{
							EventType: TimeoutDone,
							Value:     gc.Get(),
						}
						return
					}
				case <-ticker.C:
					gc.toCurrent -= time.Second
					if gc.toCurrent > 0 {
						gc.C <- Event{
							EventType: Tick,
							Value:     gc.Get(),
						}
					} else {
						go gc.done()
						return
					}

				}
			}
		}
		switch side {
		case data.TeamHome:
			if gc.homeTOs < gc.config.TimeoutsAllowed {
				gc.homeTOs++
				go timeoutRoutine()
				return
			}
		case data.TeamAway:
			if gc.awayTOs < gc.config.TimeoutsAllowed {
				gc.awayTOs++
				go timeoutRoutine()
				return
			}
		}

	}
}

func (gc *GameClock) GetTimeouts() map[string]int {
	return map[string]int{
		"home":    gc.homeTOs,
		"away":    gc.awayTOs,
		"allowed": gc.config.TimeoutsAllowed,
	}
}

// Play starts game clock at current time.
func (gc *GameClock) Play() {
	switch gc.state {
	case StatePlaying, StateDone, StateClosed:
		return
	default:
		go func() {
			gc.state = StatePlaying

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			gc.C <- Event{
				EventType: Transport,
				Value:     gc.Get(),
			}

			for {
				select {
				case <-gc.stop:
					return
				case <-ticker.C:
					gc.current -= time.Second
					if gc.current > 0 {
						gc.C <- Event{
							EventType: Tick,
							Value:     gc.Get(),
						}
					} else {
						go gc.done()
						return
					}
				}
			}
		}()
	}
}

// Pause will pause GameClock if in StatePlaying state.
func (gc *GameClock) Pause() {
	if gc.state == StatePlaying {
		gc.stop <- true
		gc.state = StatePaused
		gc.C <- Event{
			EventType: Transport,
			Value:     "",
		}
	}
	return
}

// Reset sets current clock duration to PeriodLength in cfg,
// or OtDuration if period is greater than PeriodCount in cfg.
// Will return with no action if GameClock state is StatePlaying.
func (gc *GameClock) Reset() {
	if gc.state == StatePlaying || gc.state == StateClosed {
		return
	}
	if gc.period <= gc.config.PeriodCount {
		gc.current = gc.config.PeriodLength
	} else {
		gc.current = gc.config.OtDuration
	}
	gc.state = StateFresh

	gc.C <- Event{
		EventType: ClockSet,
		Value:     gc.Get(),
	}
	return
}

// Set takes a ClockDuration and assigns its time.Duration as current GameClock time.
func (gc *GameClock) Set(dur Duration) {
	if gc.state == StatePlaying || gc.state == StateClosed {
		return
	}
	duration, err := dur.ToDuration()
	if err != nil {
		return
	}

	gc.current = duration
	gc.state = StateFresh

	gc.C <- Event{
		EventType: ClockSet,
		Value:     gc.Get(),
	}
	return
}

// Adjust takes a ClockDuration and bool and adds time.Duration from ClockDuration to current
// GameClock time if bool is true, or subtracts if bool is false.
func (gc *GameClock) Adjust(dur Duration, add bool) {
	if gc.state == StatePlaying || gc.state == StateClosed {
		return
	}
	duration, err := dur.ToDuration()
	if err != nil {
		return
	}

	if add {
		gc.current += duration
	} else {
		gc.current -= duration
	}
	gc.state = StateFresh

	gc.C <- Event{
		EventType: ClockSet,
		Value:     gc.Get(),
	}
	return
}

// Get returns a string in format of "MM:SS" with current GameClock time
func (gc *GameClock) Get() string {
	if gc.state == StateClosed {
		return ""
	}

	var current time.Duration
	if gc.state == StateTimeout {
		current = gc.toCurrent
	} else {
		current = gc.current
	}

	mins := int(math.Floor(current.Minutes()))
	minsDuration := time.Duration(mins) * time.Minute
	secs := int(math.Round((current - minsDuration).Seconds()))
	var padMin string
	var padSec string
	switch {
	case mins < 10:
		padMin = "0"
	default:
		padMin = ""
	}
	switch {
	case secs < 10:
		padSec = "0"
	default:
		padSec = ""
	}
	str := fmt.Sprintf(`%s%d:%s%d`, padMin, mins, padSec, secs)
	return str
}

// ChangePeriod sets the current GameClock period.
// Period can only be changed on a GameClock with StateFresh or StateDone state
func (gc *GameClock) ChangePeriod(add int64) {
	if gc.state == StatePlaying || gc.state == StatePaused || gc.state == StateClosed {
		return
	}
	if gc.period+add <= 0 {
		return
	}
	gc.period += add
	gc.current = gc.config.PeriodLength
	gc.C <- Event{
		EventType: PeriodSet,
		Value:     fmt.Sprintf("%d/%d", gc.period, gc.config.PeriodCount),
	}
}

// GetPeriod returns current GameClock period.
func (gc *GameClock) GetPeriod() int64 {
	if gc.state == StatePlaying {
		return 0
	}
	return gc.period
}

func (gc *GameClock) Close() {
	if gc.state == StatePlaying {
		return
	}
	defer close(gc.C)
	defer close(gc.stop)
	defer close(gc.Controller)
	gc.state = StateClosed
}

// StateDone is called when GameClock current is 0 or less.
func (gc *GameClock) done() {
	if gc.current > 0 {
		gc.state = StatePaused
		gc.toCurrent = gc.config.TimeoutDuration
		gc.C <- Event{
			EventType: TimeoutDone,
			Value:     gc.Get(),
		}
		return // Will be greater than 0 if StateTimeout timer was used
	}
	gc.current = 0
	gc.state = StateDone
	gc.C <- Event{
		EventType: Done,
		Value:     "",
	}
	return
}

type Config struct {
	PeriodLength    time.Duration
	PeriodCount     int64
	OtDuration      time.Duration
	TimeoutDuration time.Duration
	TimeoutsAllowed int
}

type Control int

const (
	Play Control = iota
	Pause
	Reset
	AddMin
	SubtractMin
	AddSec
	SubtractSec
	AddPeriod
	SubtractPeriod
	CallTimeoutHome
	CallTimeoutAway
	EndTimeout
)

type EventType int

const (
	Tick EventType = iota
	Transport
	Done
	ClockSet
	PeriodSet
	Timeout
	TimeoutDone
)

type Event struct {
	EventType
	Value string
}

func NewGameClock(cfg Config) *GameClock {
	clock := &GameClock{
		current:    cfg.PeriodLength,
		toCurrent:  cfg.TimeoutDuration,
		state:      StateFresh,
		period:     1,
		C:          make(chan Event),
		config:     cfg,
		stop:       make(chan bool),
		Controller: make(chan Control),
		homeTOs:    0,
		awayTOs:    0,
	}

	go clock.run()

	return clock
}
