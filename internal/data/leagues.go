package data

import (
	"ScoreTableApi/internal/pins"
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"time"
)

type League struct {
	ID        int64     `json:"-"`
	Pin       pins.Pin  `json:"pin"`
	UserID    int64     `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"-"`
	IsActive  bool      `json:"is_active"`
}

func (l *League) Validate(v *validator.Validator) {
	v.Check(len(l.Name) <= 20, "name", "must be 20 characters or less")
}

type LeagueModel struct {
	db *sql.DB
}

func (m *LeagueModel) Create(l *League) (*League, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	pin, err := helperModels.Pins.New(pins.PinScopeLeagues, tx, ctx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	stmt := `
	INSERT INTO leagues (name, pin_id, user_id)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, version, is_active
	`

	tx.QueryRowContext(ctx, stmt, args...)

}
