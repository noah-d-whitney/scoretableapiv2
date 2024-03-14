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

type Player struct {
	ID         int64     `json:"id"`
	PinId      pins.Pin  `json:"pin_id"`
	UserId     int64     `json:"user_id"`
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	pin, err := helperModels.Pins.New(pins.PinScopePlayers, tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	player.PinId = *pin

	stmt := `
		INSERT INTO players (user_id, pin_id, first_name, last_name, pref_number, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, created_at, version, is_active`

	args := []any{
		player.UserId,
		player.PinId.ID,
		player.FirstName,
		player.LastName,
		player.PrefNumber,
	}

	err = tx.QueryRowContext(ctx, stmt, args...).Scan(&player.ID, &player.CreatedAt,
		&player.Version, &player.IsActive)
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

func (m *PlayerModel) Get(id int64) (*Player, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	stmt := `
		SELECT id, first_name, last_name, pref_number, created_at, version, is_active
		FROM players
		WHERE id = $1`

	var player Player
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, stmt, id).Scan(
		&player.ID,
		&player.FirstName,
		&player.LastName,
		&player.PrefNumber,
		&player.CreatedAt,
		&player.Version,
		&player.IsActive,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &player, nil
}

func (m *PlayerModel) GetAll(name string, filters Filters) ([]*Player, Metadata, error) {
	stmt := fmt.Sprintf(`
		SELECT count(*) OVER(), id, first_name, last_name, pref_number, created_at, version, 
			is_active
		FROM players
		WHERE (to_tsvector('simple', first_name) @@ plainto_tsquery('simple', $1) 
			OR to_tsvector('simple', last_name) @@ plainto_tsquery('simple', $1)
			OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{name, filters.limit(), filters.offset()}

	rows, err := m.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	players := []*Player{}
	for rows.Next() {
		var player Player
		err := rows.Scan(
			&totalRecords,
			&player.ID,
			&player.FirstName,
			&player.LastName,
			&player.PrefNumber,
			&player.CreatedAt,
			&player.Version,
			&player.IsActive,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		players = append(players, &player)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return players, metadata, nil
}

func (m *PlayerModel) Update(player *Player) error {
	stmt := `
		UPDATE players
		SET first_name = $1, last_name = $2, pref_number = $3, is_active = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []any{
		player.FirstName,
		player.LastName,
		player.PrefNumber,
		player.IsActive,
		player.ID,
		player.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, stmt, args...).Scan(&player.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m *PlayerModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	stmt := `
		DELETE FROM players
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
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
