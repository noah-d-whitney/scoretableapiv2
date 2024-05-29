package data

import (
	"ScoreTableApi/internal/pins"
	"ScoreTableApi/internal/validator"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// TODO: implement allowedkeepers table & model

type Game struct {
	ID             int64         `json:"-"`
	UserID         int64         `json:"-"`
	PinID          pins.Pin      `json:"pin_id"`
	CreatedAt      time.Time     `json:"-"`
	Version        int64         `json:"-"`
	Status         GameStatus    `json:"status"`
	DateTime       time.Time     `json:"date_time"`
	TeamSize       int64         `json:"team_size"`
	Type           GameType      `json:"type"`
	PeriodLength   *PeriodLength `json:"period_length,omitempty"`
	PeriodCount    *int64        `json:"period_count,omitempty"`
	ScoreTarget    *int64        `json:"score_target,omitempty"`
	HomeTeamPin    *string       `json:"home_team_pin,omitempty"`
	AwayTeamPin    *string       `json:"away_team_pin,omitempty"`
	HomePlayerPins []string      `json:"-"`
	AwayPlayerPins []string      `json:"-"`
	Teams          struct {
		Home *Team `json:"home,omitempty"`
		Away *Team `json:"away,omitempty"`
	} `json:"teams,omitempty"`
}

func (g *Game) GetPlayerPins() (homeTeamPins, awayTeamPins []string) {
	if g.Teams.Home == nil && g.Teams.Away == nil {
		return nil, nil
	}

	homeTeamPins = make([]string, 0)
	for _, p := range g.Teams.Home.Players {
		homeTeamPins = append(homeTeamPins, p.PinId.Pin)
	}

	awayTeamPins = make([]string, 0)
	for _, p := range g.Teams.Away.Players {
		awayTeamPins = append(awayTeamPins, p.PinId.Pin)
	}

	return homeTeamPins, awayTeamPins
}

type GameDto struct {
	DateTime     *time.Time    `json:"date_time"`
	TeamSize     *int64        `json:"team_size"`
	Type         *GameType     `json:"type"`
	PeriodLength *PeriodLength `json:"period_length"`
	PeriodCount  *int64        `json:"period_count"`
	ScoreTarget  *int64        `json:"score_target"`
	HomeTeamPin  *string       `json:"home_team_pin"`
	AwayTeamPin  *string       `json:"away_team_pin"`
}

func (dto GameDto) validate(v *validator.Validator) {
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
	}

	if dto.TeamSize != nil {
		v.Check(*dto.TeamSize > 0, "team_size", "must be greater than 0")
		v.Check(*dto.TeamSize <= 5, "team_size", "must be 5 or less")
	}

	if dto.Type != nil {
		v.Check(*dto.Type == GameTypeTimed || *dto.Type == GameTypeTarget, "type",
			fmt.Sprintf(`Must be one of the following: "%s", "%s"`, GameTypeTimed,
				GameTypeTarget))

		if *dto.Type == GameTypeTimed {
			v.Check(dto.PeriodCount != nil, "period_count", "must be provided for timed game")
			v.Check(dto.PeriodLength != nil, "period_length", "must be provided for timed game")
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
			v.Check(dto.PeriodLength == nil, "period_length",
				"cannot be provided for a target game")
			if !v.Valid() {
				return
			}

			v.Check(*dto.ScoreTarget > 0, "score_target", "must be greater than 0")
			v.Check(*dto.ScoreTarget <= 100, "score_target", "must be 100 or less")
		}
	} else {
		v.Check(dto.ScoreTarget == nil, "score_target", "cannot be provided without type field")
		v.Check(dto.PeriodCount == nil, "period_count", "cannot be provided without type field")
		v.Check(dto.PeriodLength == nil, "period_length", "cannot be provided without type field")
	}
}

func (dto GameDto) Merge(v *validator.Validator, g *Game) {
	dto.validate(v)
	if !v.Valid() {
		return
	}

	if dto.DateTime != nil {
		if *dto.DateTime == g.DateTime {
			v.AddError("date_time", "cannot be old value")
		} else {
			g.DateTime = *dto.DateTime
		}
	}
	if dto.TeamSize != nil {
		if *dto.TeamSize == g.TeamSize {
			v.AddError("team_size", "cannot be old value")
		} else {
			g.TeamSize = *dto.TeamSize
		}
	}
	if dto.PeriodLength != nil {
		if dto.PeriodLength == g.PeriodLength {
			v.AddError("period_length", "cannot be old value")
		} else {
			g.PeriodLength = dto.PeriodLength
		}
	}
	if dto.PeriodCount != nil {
		if dto.PeriodCount == g.PeriodCount {
			v.AddError("period_count", "cannot be old value")
		} else {
			g.PeriodCount = dto.PeriodCount
		}
	}
	if dto.ScoreTarget != nil {
		if dto.ScoreTarget == g.ScoreTarget {
			v.AddError("score_target", "cannot be old value")
		} else {
			g.ScoreTarget = dto.ScoreTarget
		}
	}
	if dto.HomeTeamPin != nil {
		g.HomeTeamPin = dto.HomeTeamPin
	}
	if dto.AwayTeamPin != nil {
		g.AwayTeamPin = dto.AwayTeamPin
	}

	return
}

func (dto GameDto) Convert(v *validator.Validator) *Game {
	if dto.DateTime == nil {
		v.AddError("date_time", "must be provided")
	}
	if dto.TeamSize == nil {
		v.AddError("team_size", "must be provided")
	}
	if dto.Type == nil {
		v.AddError("type", "must be provided")
	}
	if !v.Valid() {
		return nil
	}

	dto.validate(v)
	if !v.Valid() {
		return nil
	}

	var game *Game
	game.DateTime = *dto.DateTime
	game.TeamSize = *dto.TeamSize
	game.Type = *dto.Type
	if dto.PeriodLength != nil {
		game.PeriodLength = dto.PeriodLength
	}
	if dto.PeriodCount != nil {
		game.PeriodCount = dto.PeriodCount
	}
	if dto.ScoreTarget != nil {
		game.ScoreTarget = dto.ScoreTarget
	}
	if dto.HomeTeamPin != nil {
		game.HomeTeamPin = dto.HomeTeamPin
	}
	if dto.AwayTeamPin != nil {
		game.AwayTeamPin = dto.AwayTeamPin
	}

	return game
}

type GameModel struct {
	db *sql.DB
}

// TODO refactor PeriodLength JSON mar/um

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
