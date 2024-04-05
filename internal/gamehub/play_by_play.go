package gamehub

import "ScoreTableApi/internal/data"

//type PlayEngine struct {
//	in  chan GameEvent
//	out chan Play
//}
//
//func (e *PlayEngine) Run() {
//	for {
//		select {
//		case <-in
//		}
//	}
//}

type PlayerPlay struct {
	period *int64
	time   *string
	team   struct {
		name string
		pin  string
		side data.GameTeamSide
	}
	player struct {
		name   string
		pin    string
		number int
	}
	description string
}
