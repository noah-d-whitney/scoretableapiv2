package data

import (
	"ScoreTableApi/internal/pins"
	"context"
	"database/sql"
	"errors"
	"time"
)

func (m *GameModel) Delete(userID int64, pin string) error {
	stmt := `
		DELETE FROM games
		USING pins
		WHERE user_id = $1 AND pin = $2
		RETURNING pin_id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var pinID int64
	err = tx.QueryRowContext(ctx, stmt, userID, pin).Scan(&pinID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	err = helperModels.Pins.Delete(pinID, pins.PinScopeGames, tx, ctx)
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
