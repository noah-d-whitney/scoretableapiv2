package data

import (
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"time"
)

type Player struct {
	ID         int64     `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	PrefNumber int       `json:"pref_number"`
	CreatedAt  time.Time `json:"-"`
	Version    int32     `json:"-"`
	IsActive   bool      `json:"active"`
}

type PlayerModel struct {
	db *sql.DB
}

func (m *PlayerModel) Insert(player *Player) error {
	stmt := `
		INSERT INTO players (first_name, last_name, pref_number, is_active)
		VALUES ($1, $2, $3, true)
		RETURNING id, created_at, version, is_active`

	args := []any{
		player.FirstName,
		player.LastName,
		player.PrefNumber,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.db.QueryRowContext(ctx, stmt, args...).Scan(&player.ID, &player.CreatedAt,
		&player.Version, &player.IsActive)
}

func ValidatePlayer(v *validator.Validator, player *Player) {
	v.Check(player.FirstName != "", "first_name", "must be provided")
	v.Check(len(player.FirstName) > 1, "first_name", "must be greater than 1 character")
	v.Check(len(player.FirstName) <= 20, "first_name", "must be 20 character or less")

	v.Check(player.LastName != "", "last_name", "must be provided")
	v.Check(len(player.LastName) > 1, "last_name", "must be greater than 1 character")
	v.Check(len(player.LastName) <= 20, "last_name", "must be 20 characters or less")

	v.Check(player.PrefNumber >= 0, "pref_number", "must be 0 or greater")
	v.Check(player.PrefNumber < 100, "pref_number", "must be less than 100")
}
