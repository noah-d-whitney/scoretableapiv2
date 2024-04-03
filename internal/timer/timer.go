package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	strings2 "strings"
	"sync"
	"time"
)

//start: start timer count and print every second time to string
//pause: pause timer count and print pause message with current time
//10 secs left: start print every 1/10 of sec until 0
//0 secs: print done, increment period/reset for next

// need : print func (print interval) -> c channel. reset (
//) checks periods and does nothing if current = count, else iuncrement current and reset current.
//

var ErrInvalidDuration = errors.New("invalid timer duration string")

// ClockDuration represents a string in the format "MM:SS"
type ClockDuration string

// ToDuration converts string from format "MM:SS" to a time.Duration
func (cd ClockDuration) ToDuration() (time.Duration, error) {
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

type state int

const (
	fresh state = iota
	playing
	paused
	done
)

// GameClock keeps current game time and period. Sends string every second on C with current time
// when GameClock is running.
type GameClock struct {
	current time.Duration
	C       chan string
	state   state
	period  int64
	config  GameClockConfig
	stop    chan bool
	done    chan bool
	wg      sync.WaitGroup
}

// Play starts game clock at current time.
func (gc *GameClock) Play() {
	switch gc.state {
	case playing:
		return
	default:
		go func() {
			gc.wg.Add(1)
			defer gc.wg.Done()

			gc.state = playing

			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()
			go gc.sendAtInterval(time.Second)
			go gc.listenDone(gc.done)

			fmt.Printf("Clock started @ %s\n", gc.Get())

			for {
				select {
				case pause := <-gc.stop:
					gc.stop <- pause // send stop signal again to terminate ticker
					return
				case <-ticker.C:
					gc.current -= time.Millisecond
					if gc.current <= 0 {
						gc.done <- true
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
		gc.wg.Wait()
		gc.state = paused
		fmt.Printf("Clock paused\n")
	}
	return
}

// Reset sets current clock duration to PeriodLength in cfg,
// or OtDuration if period is greater than PeriodCount in cfg.
// Will return with no action if GameClock state is playing.
func (gc *GameClock) Reset() {
	if gc.state == playing {
		return
	}
	if gc.period <= gc.config.PeriodCount {
		gc.current = gc.config.PeriodLength
	} else {
		gc.current = gc.config.OtDuration
	}
	gc.state = fresh

	fmt.Printf("Clock reset to %s\n", gc.Get())
	return
}

// Set takes a ClockDuration and assigns its time.Duration as current GameClock time.
func (gc *GameClock) Set(dur ClockDuration) {
	if gc.state == playing {
		return
	}
	duration, err := dur.ToDuration()
	if err != nil {
		return
	}

	gc.current = duration
	gc.state = fresh

	fmt.Printf("Clock set to %s\n", gc.Get())
	return
}

// Adjust takes a ClockDuration and bool and adds time.Duration from ClockDuration to current
// GameClock time if bool is true, or subtracts if bool is false.
func (gc *GameClock) Adjust(dur ClockDuration, add bool) {
	if gc.state == playing {
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

	fmt.Printf("Clock set to %s\n", gc.Get())
	return
}

// Get returns a string in format of "MM:SS" with current GameClock time
func (gc *GameClock) Get() string {
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
	str := fmt.Sprintf(`"%s%d:%s%d"`, padMin, mins, padSec, secs)
	return str
}

// ChangePeriod sets the current GameClock period.
// Period can only be changed on a GameClock with fresh or done state
func (gc *GameClock) ChangePeriod(add int64) {
	if gc.state == playing || gc.state == paused {
		return
	}
	if gc.period+add <= 0 {
		return
	}
	gc.period += add
	gc.Reset()
	fmt.Printf("Next Period Started\n")
}

// GetPeriod returns current GameClock period.
func (gc *GameClock) GetPeriod() int64 {
	return gc.period
}

func (gc *GameClock) sendAtInterval(interval time.Duration) {
	gc.wg.Add(1)
	defer gc.wg.Done()
	for {
		select {
		default:
			time.Sleep(interval)
			fmt.Printf("%s", gc.Get())
		case <-gc.stop:
			return
		}
	}
}

// listenDone listens for bool from doneChan channel and performs tasks
func (gc *GameClock) listenDone(doneChan chan bool) {
	<-doneChan
	gc.stop <- true
	gc.wg.Wait()
	gc.current = 0
	gc.state = done
	fmt.Printf("Clock done\n")
	return
}

type GameClockConfig struct {
	PeriodLength time.Duration
	PeriodCount  int64
	OtDuration   time.Duration
}

func newGameClock(cfg GameClockConfig) *GameClock {
	clock := &GameClock{
		current: cfg.PeriodLength,
		state:   fresh,
		period:  1,
		C:       make(chan string),
		config:  cfg,
		stop:    make(chan bool),
		done:    make(chan bool),
		wg:      sync.WaitGroup{},
	}

	return clock
}

func main() {
	clock := newGameClock(GameClockConfig{
		PeriodLength: 10 * time.Second,
		PeriodCount:  4,
		OtDuration:   5 * time.Second,
	})

	go func() {
		time.Sleep(3 * time.Second)
		clock.Reset()
	}()

	clock.Play()
	<-time.NewTimer(5 * time.Minute).C
}
