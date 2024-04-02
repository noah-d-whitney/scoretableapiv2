package main

import (
	"fmt"
	"math"
	"time"
)

//start: start timer count and print every second time to string
//pause: pause timer count and print pause message with current time
//10 secs left: start print every 1/10 of sec until 0
//0 secs: print done, increment period/reset for next

// need : print func (print interval) -> c channel. reset (
//) checks periods and does nothing if current = count, else iuncrement current and reset current.
//

type gameClock struct {
	current time.Duration
	length  time.Duration
	c       chan string
	periods struct {
		current int
		count   int64
	}
	stop chan bool
	done chan bool
}

func (gc *gameClock) play() {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	go gc.sendAtInterval(time.Second)
	fmt.Printf("Clock Started @ %d\n", gc.current)
	for {
		select {
		case <-gc.stop:
			fmt.Printf("Clock Paused @ %d\n", gc.current)
			return
		case <-ticker.C:
			gc.current -= time.Millisecond
			if gc.current <= 0 {
				fmt.Printf("Clock done!\n")
				gc.done <- true
				return
			}
		}
	}
}

func (gc *gameClock) pause() {
	gc.stop <- true
}

func (gc *gameClock) get() string {
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

func (gc *gameClock) sendAtInterval(interval time.Duration) {
	for {
		select {
		case <-gc.stop:
			println("Stopped from interval sender")
			return
		default:
			time.Sleep(interval)
			fmt.Printf("%s\n", gc.get())
		}
	}
}

func newGameClock(periodLength time.Duration, periodCount int64) *gameClock {
	clock := &gameClock{
		current: periodLength,
		length:  periodLength,
		c:       make(chan string),
		periods: struct {
			current int
			count   int64
		}{
			current: 1,
			count:   periodCount,
		},
		stop: make(chan bool),
		done: make(chan bool),
	}

	return clock
}

func main() {
	clock := newGameClock(30*time.Second, 4)
	go func() {
		time.Sleep(8*time.Second + 400*time.Millisecond)
		clock.pause()
		time.Sleep(3 * time.Second)
		clock.play()
		time.Sleep(7 * time.Second)
		clock.pause()
		return
	}()
	clock.play()
	<-clock.done
}
