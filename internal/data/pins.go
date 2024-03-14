package data

import (
	"context"
	"database/sql"
	"errors"
	"math/rand/v2"
	"time"
)

var (
	errDuplicatePin = errors.New("duplicate pin")
	SCOPE_TEAMS     = "teams"
)

type Pin struct {
	ID    int64
	Pin   string
	Scope string
}

type PinModel struct {
	db *sql.DB
}

func (m *PinModel) insert(pin *Pin) error {
	stmt := `
		INSERT INTO pins (pin, scope)
		VALUES ($1, $2)
		RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, stmt, pin.Pin, pin.Scope).Scan(&pin.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "pins_pin_key"`:
			return errDuplicatePin
		default:
			return err
		}
	}

	return nil
}

func (m *PinModel) New(scope string) (*Pin, error) {
	pinString := m.generatePin()
	pin := &Pin{
		Pin:   pinString,
		Scope: scope,
	}

	err := m.insert(pin)
	if err != nil {
		switch {
		case errors.Is(err, errDuplicatePin):
			return m.New(scope)
		default:
			return nil, err
		}
	}

	return pin, nil
}

var (
	pinLength   = 6
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
)

func (m *PinModel) generatePin() string {
	b := make([]rune, pinLength)
	for i := range b {
		b[i] = letterRunes[rand.IntN(len(letterRunes))]
	}
	return string(b)
}
