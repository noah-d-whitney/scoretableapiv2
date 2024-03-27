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

func (g Game) ToDto() GameDto {
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

type UpdateGameDto struct {
	DateTime     *time.Time   `json:"date_time"`
	TeamSize     *int64       `json:"team_size"`
	Type         *GameType    `json:"type"`
	PeriodLength PeriodLength `json:"period_length"`
	PeriodCount  *int64       `json:"period_count"`
	ScoreTarget  *int64       `json:"score_target"`
	HomeTeamPin  *string      `json:"home_team_pin"`
	AwayTeamPin  *string      `json:"away_team_pin"`
}

func (dto UpdateGameDto) validate(v *validator.Validator) {
	//v.Check(dto.DateTime != nil, "date_time", "must be provided")
	//v.Check(dto.TeamSize != nil, "team_size", "must be provided")
	//v.Check(dto.Type != nil, "type", "must be provided")
	if dto.HomeTeamPin != nil || dto.AwayTeamPin != nil {
		if dto.HomeTeamPin == dto.AwayTeamPin {
			v.AddError("home_team_pin", "cannot match away team")
			v.AddError("away_team_pin", "cannot match home team")
		}
	}
	if !v.Valid() {
		return
	}

	if dto.DateTime != nil {
		v.Check(dto.DateTime.After(time.Now()), "date_time", "must be in the future")
		v.Check(*dto.TeamSize > 0, "team_size", "must be greater than 0")
		v.Check(*dto.TeamSize <= 5, "team_size", "must be 5 or less")
		v.Check(*dto.Type == GameTypeTimed || *dto.Type == GameTypeTarget, "type",
			fmt.Sprintf(`Must be one of the following: "%s", "%s"`, GameTypeTimed,
				GameTypeTarget))
	}

	if dto.Type != nil {
		if *dto.Type == GameTypeTimed {
			v.Check(dto.PeriodCount != nil, "period_count", "must be provided for timed game")
			v.Check(dto.PeriodLength != 0, "period_length", "must be provided for timed game")
			v.Check(dto.ScoreTarget == nil, "score_target", "cannot be provided for a timed game")
			if !v.Valid() {
				return
			}

			v.Check(dto.PeriodLength.Duration() <= 30*time.Minute, "period_count",
				"must be 30 minutes or less")

			v.Check(*dto.PeriodCount > 0, "period_count", "must be greater than 0")
			v.Check(*dto.PeriodCount <= 4, "period_count", "must be 4 or less")
		}

		if *dto.Type == GameTypeTarget {
			v.Check(dto.ScoreTarget != nil, "score_target", "must be provided for target game")
			v.Check(dto.PeriodCount == nil, "period_count", "cannot be provided for a target game")
			v.Check(dto.PeriodLength == 0, "period_length", "cannot be provided for a target game")
			if !v.Valid() {
				return
			}

			v.Check(*dto.ScoreTarget > 0, "score_target", "must be greater than 0")
			v.Check(*dto.ScoreTarget <= 100, "score_target", "must be 100 or less")
		}
	}
}

func (dto UpdateGameDto) Convert(v *validator.Validator) (*Game, GameAux) {
	dto.validate(v)
	if !v.Valid() {
		return nil, GameAux{}
	}

	game := &Game{
		DateTime:     dto.DateTime,
		TeamSize:     dto.TeamSize,
		Type:         dto.Type,
		PeriodLength: dto.PeriodLength,
		PeriodCount:  dto.PeriodCount,
		ScoreTarget:  dto.ScoreTarget,
	}

	aux := GameAux{
		HomeTeamPin: dto.HomeTeamPin,
		AwayTeamPin: dto.AwayTeamPin,
	}

	return game, aux
}

func (dto UpdateGameDto) Merge(v *validator.Validator, g *Game) GameAux {
	dto.validate(v)
	if !v.Valid() {
		return GameAux{}
	}

	g.DateTime = dto.DateTime
	g.TeamSize = dto.TeamSize
	g.PeriodLength = dto.PeriodLength
	g.PeriodCount = dto.PeriodCount
	g.ScoreTarget = dto.ScoreTarget

	aux := GameAux{
		HomeTeamPin: dto.HomeTeamPin,
		AwayTeamPin: dto.AwayTeamPin,
	}

	return aux
}

type GameAux struct {
	HomeTeamPin *string
	AwayTeamPin *string
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
