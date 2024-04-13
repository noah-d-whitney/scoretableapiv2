package data

import (
	"ScoreTableApi/internal/pins"
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// todo change player numbers to ptrs to allow #0
type Player struct {
	ID         int64     `json:"-"`
	PinId      pins.Pin  `json:"pin"`
	UserId     int64     `json:"-"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	PrefNumber int       `json:"pref_number,omitempty"`
	CreatedAt  time.Time `json:"-"`
	Version    int32     `json:"-"`
	IsActive   bool      `json:"is_active"`
	Number     int       `json:"number,omitempty"`
	LineupPos  *int      `json:"lineup_pos,omitempty"`
}

// TODO add player status (injured, unavailable, etc)
// and make game unavailable to start until lineup is changed
// TODO make views in psql table for common queries (teamsize, is_active, team player number,
// game type,  etc)

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

	pin, err := helperModels.Pins.New(pins.PinScopePlayers, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	player.PinId = *pin

	stmt := `
		INSERT INTO players (user_id, pin_id, first_name, last_name, pref_number)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version`

	args := []any{
		player.UserId,
		player.PinId.ID,
		player.FirstName,
		player.LastName,
		player.PrefNumber,
	}

	err = tx.QueryRowContext(ctx, stmt, args...).Scan(&player.ID, &player.CreatedAt,
		&player.Version)
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

func (m *PlayerModel) Get(userId int64, pin string) (*Player, error) {
	stmt := `
		SELECT pins.id, pins.pin, pins.scope, players.id, players.first_name, players.last_name, 
			players.pref_number, players.created_at, players.version, (
				SELECT count(*)::int::bool
					FROM teams_players
					JOIN teams ON teams_players.team_id = teams.id
					WHERE player_id = players.id 
						AND lineup_number IS NOT NULL 
						AND teams.is_active IS NOT FALSE)
		FROM pins
		JOIN players ON pins.id = players.pin_id
		WHERE pins.pin = $1 AND players.user_id = $2 AND pins.scope = $3`

	args := []any{pin, userId, pins.PinScopePlayers}

	var player Player
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, stmt, args...).Scan(
		&player.PinId.ID,
		&player.PinId.Pin,
		&player.PinId.Scope,
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

func (m *PlayerModel) GetList(userID int64, pins []string) ([]*Player, error) {
	stmt := `
        SELECT pins.id, pins.pin, pins.scope, players.id, players.first_name, players.last_name, players.pref_number,
            players.created_at, players.version, (
				SELECT count(*)::int::bool
					FROM teams_players
					WHERE player_id = players.id)
        FROM players
        INNER JOIN pins ON players.pin_id = pins.id
        WHERE players.user_id = $1 AND pins.pin = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.db.QueryContext(ctx, stmt, userID, pq.Array(pins))
	if err != nil {
		return nil, err
	}

	var players []*Player
	for rows.Next() {
		var player Player
		err := rows.Scan(
			&player.PinId.ID,
			&player.PinId.Pin,
			&player.PinId.Scope,
			&player.ID,
			&player.FirstName,
			&player.LastName,
			&player.PrefNumber,
			&player.CreatedAt,
			&player.Version,
			&player.IsActive,
		)
		if err != nil {
			return nil, err
		}

		players = append(players, &player)
	}

	return players, nil
}

func (m *PlayerModel) GetAll(userID int64, name string, filters Filters) ([]*Player, Metadata,
	error,
) {
	stmt := fmt.Sprintf(`
		SELECT count(*) OVER(), pins.id, pins.pin, pins.scope, players.id, players.first_name, players.last_name, 
			players.pref_number, players.created_at, players.version, (
				SELECT count(*)::int::bool
					FROM teams_players
					WHERE player_id = players.id)
		FROM players
		INNER JOIN pins ON players.pin_id = pins.id
		WHERE (user_id = $1 AND to_tsvector('simple', first_name) @@ plainto_tsquery('simple', $2) 
			OR to_tsvector('simple', last_name) @@ plainto_tsquery('simple', $2)
			OR $2 = '')
		ORDER BY %s %s, players.id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	args := []any{userID, name, filters.limit(), filters.offset()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

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
			&player.PinId.ID,
			&player.PinId.Pin,
			&player.PinId.Scope,
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
		SET first_name = $1, last_name = $2, pref_number = $3, version = version + 1
		WHERE user_id = $5 AND id = $6 AND version = $7
		RETURNING version`

	args := []any{
		player.FirstName,
		player.LastName,
		player.PrefNumber,
		player.UserId,
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

// todo fix team lineups when unassigning player from team
func (m *PlayerModel) Delete(userID int64, pin string) error {
	stmt := `
		DELETE FROM players
		USING pins
		WHERE players.user_id = $1
			AND pins.pin = $2
			AND players.pin_id = pins.id
		RETURNING players.pin_id, players.id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var pinID int64
	var id int64
	err = tx.QueryRowContext(ctx, stmt, userID, pin).Scan(&pinID, &id)
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

	err = helperModels.Pins.Delete(pinID, pins.PinScopePlayers, tx, ctx)
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
