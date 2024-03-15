package data

import (
	"ScoreTableApi/internal/pins"
	"context"
	"database/sql"
)

type PinModel struct {
	db *sql.DB
}

func (m *PinModel) New(scope string, tx *sql.Tx, ctx context.Context) (*pins.Pin, error) {
	pinString := pins.GeneratePin(6)
	pin := &pins.Pin{
		Pin:   pinString,
		Scope: scope,
	}

	stmt := `
		INSERT INTO pins (pin, scope)
		VALUES ($1, $2)
		RETURNING id`

	err := tx.QueryRowContext(ctx, stmt, pin.Pin, pin.Scope).Scan(&pin.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "pins_pin_key"`:
			return m.New(scope, tx, ctx)
		default:
			return nil, err
		}
	}

	return pin, nil
}

func (m *PinModel) Delete(id int64, scope string, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		DELETE FROM pins
		WHERE id = $1 AND scope = $2`

	_, err := tx.ExecContext(ctx, stmt, id, scope)
	return err
}
