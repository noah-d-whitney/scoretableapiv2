package data

import (
	"ScoreTableApi/internal/pins"
	"ScoreTableApi/internal/validator"
	"database/sql"
	json2 "encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type Game struct {
	ID           int64        `json:"-"`
	UserID       int64        `json:"-"`
	PinID        pins.Pin     `json:"pin_id"`
	CreatedAt    time.Time    `json:"-"`
	Version      int64        `json:"-"`
	Status       GameStatus   `json:"status"`
	DateTime     *time.Time   `json:"date_time"`
	TeamSize     *int64       `json:"team_size"`
	Type         *GameType    `json:"type"`
	PeriodLength PeriodLength `json:"period_length,omitempty"`
	PeriodCount  *int64       `json:"period_count,omitempty"`
	ScoreTarget  *int64       `json:"score_target,omitempty"`
	HomeTeamPin  string       `json:"home_team_pin,omitempty"`
	AwayTeamPin  string       `json:"away_team_pin,omitempty"`
	Teams        struct {
		Home *Team `json:"home,omitempty"`
		Away *Team `json:"away,omitempty"`
	} `json:"teams,omitempty"`
}

func (g Game) ToDto() Dto {
	return GameDto{
		Pin:          g.PinID.Pin,
		Status:       g.Status,
		DateTime:     *g.DateTime,
		TeamSize:     *g.TeamSize,
		Type:         *g.Type,
		PeriodLength: g.PeriodLength,
		PeriodCount:  *g.PeriodCount,
		ScoreTarget:  *g.ScoreTarget,
		Teams: struct {
			Home *Team `json:"home,omitempty"`
			Away *Team `json:"away,omitempty"`
		}{Home: g.Teams.Home, Away: g.Teams.Away},
	}
}

type GameDto struct {
	Pin          string       `json:"pin"`
	Status       GameStatus   `json:"status"`
	DateTime     time.Time    `json:"date_time"`
	TeamSize     int64        `json:"team_size"`
	Type         GameType     `json:"type"`
	PeriodLength PeriodLength `json:"period_length,omitempty"`
	PeriodCount  int64        `json:"period_count,omitempty"`
	ScoreTarget  int64        `json:"score_target,omitempty"`
	Teams        struct {
		Home *Team `json:"home,omitempty"`
		Away *Team `json:"away,omitempty"`
	} `json:"teams,omitempty"`
}

func (dto GameDto) Convert(_ *validator.Validator) (Resource, any) {
	return nil, nil
}

func (dto GameDto) Merge(_ *validator.Validator, _ Resource) any {
	return nil
}

func (dto GameDto) validate(_ *validator.Validator) {
	return
}

type GameModel struct {
	db *sql.DB
}

// TODO refactor PeriodLength JSON mar/um

type PeriodLength time.Duration

func (pl *PeriodLength) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return &json2.UnmarshalTypeError{Field: "period_length"}
	}
	parts := strings.Split(unquoted, ":")
	if len(parts) != 2 {
		return &json2.UnmarshalTypeError{Field: "period_length"}
	}
	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return &json2.UnmarshalTypeError{Field: "period_length"}
	}
	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return &json2.UnmarshalTypeError{Field: "period_length"}
	}
	totalTime := (time.Duration(minutes) * time.Minute) + (time.Duration(seconds) * time.Second)

	*pl = PeriodLength(totalTime)
	return nil
}
func (pl *PeriodLength) MarshalJSON() ([]byte, error) {
	duration := time.Duration(*pl)
	mins := int(math.Floor(duration.Minutes()))
	minsDuration := time.Duration(mins) * time.Minute
	secs := int(math.Round((duration - minsDuration).Seconds()))
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
	json := fmt.Sprintf(`"%s%d:%s%d"`, padMin, mins, padSec, secs)
	return []byte(json), nil
}
func (pl *PeriodLength) Duration() time.Duration {
	return time.Duration(*pl)
}

type GameStatus int64

const (
	NOTSTARTED GameStatus = iota
	INPROGRESS
	FINISHED
	CANCELED
)

func (s GameStatus) MarshalJSON() ([]byte, error) {
	switch s {
	case 0:
		return []byte(`"not-started"`), nil
	case 1:
		return []byte(`"in_progress"`), nil
	case 2:
		return []byte(`"finished"`), nil
	case 3:
		return []byte(`"canceled"`), nil
	default:
		return nil, errors.New(`"invalid game status"`)
	}
}

type GameType string

const (
	GameTypeTimed  GameType = "timed"
	GameTypeTarget GameType = "target"
)

type GameTeamSide int64

const (
	TeamHome GameTeamSide = iota
	TeamAway
)

func (s GameTeamSide) String() string {
	switch s {
	case 0:
		return "home"
	case 1:
		return "away"
	default:
		return ""
	}
}
