package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

func (m *GameModel) Get(userID int64, pin string) (*Game, error) {
	stmt := `
		SELECT games_view.pin_id, games_view.pin, games_view.scope, games_view.id, 
			games_view.user_id, games_view.created_at, games_view.version, games_view.status, 
			games_view.date_time, games_view.team_size, games_view.type, games_view.period_length, 
			games_view.period_count, games_view.score_target, games_view.home_team_pin, 
			games_view.away_team_pin, games_view.home_player_pins, games_view.away_player_pins
			FROM games_view
			WHERE user_id = $1 AND pin = $2`

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
		&game.HomeTeamPin,
		&game.AwayTeamPin,
		pq.Array(&game.HomePlayerPins),
		pq.Array(&game.AwayPlayerPins),
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
