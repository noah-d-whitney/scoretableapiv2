package data

import (
	"ScoreTableApi/internal/pins"
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidPeriods  = errors.New("both length and count must be provided for periods")
	ErrInvalidGameType = errors.New("game cannot have periods and score target")
)

type Game struct {
	ID           int64      `json:"-"`
	UserID       int64      `json:"user_id"`
	PinID        pins.Pin   `json:"pin_id"`
	CreatedAt    time.Time  `json:"-"`
	Version      int64      `json:"-"`
	Status       GameStatus `json:"status"`
	DateTime     *time.Time `json:"date_time"`
	TeamSize     *int64     `json:"team_size"`
	Type         *GameType  `json:"type"`
	PeriodLength *int64     `json:"period_length,omitempty"`
	PeriodCount  *int64     `json:"period_count,omitempty"`
	ScoreTarget  *int64     `json:"score_target,omitempty"`
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

type GameModel struct {
	db *sql.DB
}

type GameType string

const (
	GameTypeTimed  GameType = "timed"
	GameTypeTarget GameType = "target"
)

func (m *GameModel) Insert(game *Game) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	pin, err := helperModels.Pins.New(pins.PinScopeGames, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	game.PinID = *pin

	stmt := `
		INSERT INTO games (user_id, pin_id, date_time, team_size, 
			period_length, period_count, score_target)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, version, status`

	args := []any{
		game.UserID,
		game.PinID.ID,
		game.DateTime,
		game.TeamSize,
		game.PeriodLength,
		game.PeriodCount,
		game.ScoreTarget,
	}

	err = tx.QueryRowContext(ctx, stmt, args...).Scan(
		&game.ID,
		&game.CreatedAt,
		&game.Version,
		&game.Status,
	)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	return nil
}

func ValidateGame(v *validator.Validator, game *Game) {
	v.Check(game.DateTime != nil, "date_time", "must be provided")
	v.Check(game.TeamSize != nil, "team_size", "must be provided")
	v.Check(game.Type != nil, "type", "must be provided")
	if !v.Valid() {
		return
	}

	v.Check(game.DateTime.After(time.Now()), "date_time", "must be in the future")
	v.Check(*game.TeamSize > 0, "team_size", "must be greater than 0")
	v.Check(*game.TeamSize <= 5, "team_size", "must be 5 or less")
	v.Check(*game.Type == GameTypeTimed || *game.Type == GameTypeTarget, "type",
		fmt.Sprintf(`Must be one of the following: "%s", "%s"`, GameTypeTimed, GameTypeTarget))

	if *game.Type == GameTypeTimed {
		v.Check(game.PeriodCount != nil, "period_count", "must be provided for timed game")
		v.Check(game.PeriodLength != nil, "period_length", "must be provided for timed game")
		v.Check(game.ScoreTarget == nil, "score_target", "cannot be provided for a timed game")
		if !v.Valid() {
			return
		}

		v.Check(*game.PeriodLength > 0, "period_length", "must be greater than 0 seconds")
		v.Check(*game.PeriodLength <= 60*30, "period_count", "must be 20 minutes or less")

		v.Check(*game.PeriodCount > 0, "period_count", "must be greater than 0")
		v.Check(*game.PeriodCount <= 4, "period_count", "must be 4 or less")
	}

	if *game.Type == GameTypeTarget {
		v.Check(game.ScoreTarget != nil, "score_target", "must be provided for target game")
		v.Check(game.PeriodCount == nil, "period_count", "cannot be provided for a target game")
		v.Check(game.PeriodLength == nil, "period_length", "cannot be provided for a target game")
		if !v.Valid() {
			return
		}

		v.Check(*game.ScoreTarget > 0, "score_target", "must be greater than 0")
		v.Check(*game.ScoreTarget <= 100, "score_target", "must be 100 or less")
	}
}
