package models

import (
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
