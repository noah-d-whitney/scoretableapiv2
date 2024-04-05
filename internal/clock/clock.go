package clock

import (
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
	fresh State = iota
	playing
	paused
	done
	closed
)

// GameClock keeps current game time and period. Sends string every second on C with current time
// when GameClock is running.
type GameClock struct {
	current    time.Duration
	C          chan Event
	Controller chan Control
	state      State
	period     int64
	config     Config
	stop       chan bool
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
			default:
			}
		}
	}
}

func (gc *GameClock) GetState() State {
	return gc.state
}

// Play starts game clock at current time.
func (gc *GameClock) Play() {
	switch gc.state {

	default:
		go func() {
			gc.state = playing

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

// Pause will pause GameClock if in playing state.
func (gc *GameClock) Pause() {
	if gc.state == playing {
		gc.stop <- true
		gc.state = paused
		gc.C <- Event{
			EventType: Transport,
			Value:     "",
		}
	}
	return
}

// Reset sets current clock duration to PeriodLength in cfg,
// or OtDuration if period is greater than PeriodCount in cfg.
// Will return with no action if GameClock state is playing.
func (gc *GameClock) Reset() {
	if gc.state == playing || gc.state == closed {
		return
	}
	if gc.period <= gc.config.PeriodCount {
		gc.current = gc.config.PeriodLength
	} else {
		gc.current = gc.config.OtDuration
	}
	gc.state = fresh

	gc.C <- Event{
		EventType: ClockSet,
		Value:     gc.Get(),
	}
	return
}

// Set takes a ClockDuration and assigns its time.Duration as current GameClock time.
func (gc *GameClock) Set(dur Duration) {
	if gc.state == playing || gc.state == closed {
		return
	}
	duration, err := dur.ToDuration()
	if err != nil {
		return
	}

	gc.current = duration
	gc.state = fresh

	gc.C <- Event{
		EventType: ClockSet,
		Value:     gc.Get(),
	}
	return
}

// Adjust takes a ClockDuration and bool and adds time.Duration from ClockDuration to current
// GameClock time if bool is true, or subtracts if bool is false.
func (gc *GameClock) Adjust(dur Duration, add bool) {
	if gc.state == playing || gc.state == closed {
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
	gc.state = fresh

	gc.C <- Event{
		EventType: ClockSet,
		Value:     gc.Get(),
	}
	return
}

// Get returns a string in format of "MM:SS" with current GameClock time
func (gc *GameClock) Get() string {
	if gc.state == closed {
		return ""
	}

	mins := int(math.Floor(gc.current.Minutes()))
	minsDuration := time.Duration(mins) * time.Minute
	secs := int(math.Round((gc.current - minsDuration).Seconds()))
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
// Period can only be changed on a GameClock with fresh or done state
func (gc *GameClock) ChangePeriod(add int64) {
	if gc.state == playing || gc.state == paused || gc.state == closed {
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
	if gc.state == playing {
		return 0
	}
	return gc.period
}

func (gc *GameClock) Close() {
	if gc.state == playing {
		return
	}
	defer close(gc.C)
	defer close(gc.stop)
	defer close(gc.Controller)
	gc.state = closed
}

// done is called when GameClock current is 0 or less.
func (gc *GameClock) done() {
	gc.current = 0
	gc.state = done
	gc.C <- Event{
		EventType: Done,
		Value:     "",
	}
	return
}

type Config struct {
	PeriodLength time.Duration
	PeriodCount  int64
	OtDuration   time.Duration
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
)

type EventType int

const (
	Tick EventType = iota
	Transport
	Done
	ClockSet
	PeriodSet
)

type Event struct {
	EventType
	Value string
}

func NewGameClock(cfg Config) *GameClock {
	clock := &GameClock{
		current:    cfg.PeriodLength,
		state:      fresh,
		period:     1,
		C:          make(chan Event),
		config:     cfg,
		stop:       make(chan bool),
		Controller: make(chan Control),
	}

	go clock.run()

	return clock
}
