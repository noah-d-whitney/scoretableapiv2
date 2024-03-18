package data

import (
	"ScoreTableApi/internal/pins"
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"errors"
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
	DateTime     time.Time  `json:"date_time"`
	TeamSize     int64      `json:"team_size"`
	Type         string     `json:"type"`
	PeriodLength int64      `json:"period_length,omitempty"`
	PeriodCount  int64      `json:"period_count,omitempty"`
	ScoreTarget  int64      `json:"score_target,omitempty"`
}

type GameStatus int64

const (
	NOTSTARTED GameStatus = iota
	INPROGRESS
	FINISHED
	CANCELED
)

type GameModel struct {
	db *sql.DB
}

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

	if game.PeriodCount == 0 && game.PeriodLength != 0 ||
		game.PeriodCount != 0 && game.PeriodLength == 0 {
		return ErrInvalidPeriods
	}
	if game.PeriodCount != 0 && game.ScoreTarget != 0 {
		return ErrInvalidGameType
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
	v.Check(game.DateTime.After(time.Now()), "date_time", "must be in the future")
}
