package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (m *GameModel) Get(userID int64, pin string) (*Game, error) {
	stmt := `
		SELECT pins.id, pins.pin, pins.scope, games.id, games.user_id, games.created_at, 
			games.version, games.status, games.date_time, games.team_size, (
				CASE WHEN games.score_target IS NULL
					THEN 'timed'
			    	WHEN games.score_target IS NOT NULL
					THEN 'target'
					ELSE ''
				END),
			games.period_length, games.period_count, games.score_target
			FROM games
			JOIN pins ON games.pin_id = pins.id
			WHERE games.user_id = $1 AND pins.pin = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	game := &Game{}
	err = tx.QueryRowContext(ctx, stmt, userID, pin).Scan(
		&game.PinID.ID,
		&game.PinID.Pin,
		&game.PinID.Scope,
		&game.ID,
		&game.UserID,
		&game.CreatedAt,
		&game.Version,
		&game.Status,
		&game.DateTime,
		&game.TeamSize,
		&game.Type,
		&game.PeriodLength,
		&game.PeriodCount,
		&game.ScoreTarget,
	)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, rollbackErr
		}
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	err = getGameTeams(game, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, rollbackErr
		}
		return nil, err
	}

	err = getGameTeamsPlayers(game, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, rollbackErr
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, rollbackErr
		}
		return nil, err
	}

	return game, nil
}
