package data

import (
	"ScoreTableApi/internal/pins"
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"
)

var (
	ErrDuplicateTeamName = errors.New("duplicate team name")
	ErrPlayerNotFound    = errors.New("player(s) not found")
	ErrTeamNotFound      = errors.New("team not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicatePlayer   = errors.New("duplicate player team assignment")
	ErrPlayerNotOnTeam   = errors.New("player not found on team")
)

type Team struct {
	ID        int64     `json:"-"`
	PinID     pins.Pin  `json:"pin"`
	UserID    int64     `json:"-"`
	Name      string    `json:"name"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"-"`
	IsActive  bool      `json:"is_active"`
	PlayerIDs []string  `json:"-"`
	Players   []*Player `json:"players,omitempty"`
}

type TeamModel struct {
	db *sql.DB
}

func (m *TeamModel) Insert(team *Team) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	pin, err := helperModels.Pins.New(pins.PinScopeTeams, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	team.PinID = *pin
	team.Players = []*Player{}

	stmt := `
		INSERT INTO teams (pin_id, user_id, name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version, is_active`

	args := []any{team.PinID.ID, team.UserID, team.Name}

	err = tx.QueryRowContext(ctx, stmt, args...).Scan(
		&team.ID,
		&team.CreatedAt,
		&team.Version,
		&team.IsActive,
	)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint`+
			` "unq_userid_team_name"`:
			return ErrDuplicateTeamName
		default:
			return err
		}
	}

	if len(team.PlayerIDs) != 0 {
		for _, p := range team.PlayerIDs {
			err := assignPlayer(team.ID, team.UserID, p, tx, ctx)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return rollbackErr
				}
				return err
			}
		}

		err = getTeamPlayers(team, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
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

func (m *TeamModel) Get(userID int64, pin string) (*Team, error) {
	stmt := `
		SELECT pins.id, pins.pin, pins.scope, teams.id, teams.user_id, teams.name, 
			teams.is_active, teams.version, teams.created_at
		FROM teams
		JOIN pins ON teams.pin_id = pins.id
		WHERE teams.user_id = $1 AND pins.pin = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	team := &Team{}
	err = tx.QueryRowContext(ctx, stmt, userID, pin).Scan(
		&team.PinID.ID,
		&team.PinID.Pin,
		&team.PinID.Scope,
		&team.ID,
		&team.UserID,
		&team.Name,
		&team.IsActive,
		&team.Version,
		&team.CreatedAt,
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

	err = getTeamPlayers(team, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, rollbackErr
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (m *TeamModel) GetAll(userID int64, name string, includes []string, filters Filters) ([]*Team,
	Metadata,
	error) {
	stmt := fmt.Sprintf(`
		SELECT count(*) OVER(), pins.id, pins.pin, pins.scope, teams.id, teams.user_id, teams.name,  
			teams.created_at, teams.version, teams.is_active, (
				SELECT count(*)
				FROM teams_players
				WHERE teams_players.user_id = $1 AND teams_players.team_id = teams.id)
		FROM teams
		INNER JOIN pins ON teams.pin_id = pins.id
		WHERE (user_id = $1 AND to_tsvector('simple', teams.name) @@ plainto_tsquery('simple', $2) 
			OR $2 = '')
		ORDER BY %s %s, teams.id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	args := []any{userID, name, filters.limit(), filters.offset()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, Metadata{}, rollbackErr
		}
		return nil, Metadata{}, err
	}

	rows, err := tx.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	teams := []*Team{}
	for rows.Next() {
		var team Team
		err := rows.Scan(
			&totalRecords,
			&team.PinID.ID,
			&team.PinID.Pin,
			&team.PinID.Scope,
			&team.ID,
			&team.UserID,
			&team.Name,
			&team.CreatedAt,
			&team.Version,
			&team.IsActive,
			&team.Size,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		team.Players = []*Player{}
		teams = append(teams, &team)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	if slices.Contains(includes, "players") {
		for _, team := range teams {
			err := getTeamPlayers(team, tx, ctx)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return nil, Metadata{}, rollbackErr
				}
				return nil, Metadata{}, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, Metadata{}, rollbackErr
		}
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return teams, metadata, nil
}

func (m *TeamModel) Delete(userID int64, pin string) error {
	stmt := `
		DELETE FROM teams
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
	err = tx.QueryRowContext(ctx, stmt, pin, userID).Scan(&pinID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	err = helperModels.Pins.Delete(pinID, pins.PinScopeTeams, tx, ctx)
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

func (m *TeamModel) Update(team *Team) error {
	stmt := `
		UPDATE teams
		SET name = $1, is_active = $2, version = version + 1
		WHERE teams.user_id = $3
			AND teams.id = $4
			AND teams.version = $5
		RETURNING version`

	args := []any{team.Name, team.IsActive, team.UserID, team.ID, team.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx, stmt, args...).Scan(&team.Version)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	assign, unassign := parsePinList(team.PlayerIDs)

	for _, pin := range assign {
		err := assignPlayer(team.ID, team.UserID, pin, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	for _, pin := range unassign {
		err := unassignPlayer(team.ID, team.UserID, pin, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	err = getTeamPlayers(team, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	team.Size = len(team.Players)

	err = tx.Commit()
	if err != nil {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	return nil
}

// TODO assign lineup #

func assignPlayer(teamID, userID int64, playerPin string, tx *sql.Tx, ctx context.Context) error {
	getStmt := `
		SELECT players.id
		FROM players
		JOIN public.pins ON pins.id = players.pin_id
		WHERE pins.pin = $1 AND players.user_id = $2`

	var playerID string
	err := tx.QueryRowContext(ctx, getStmt, playerPin, userID).Scan(&playerID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrPlayerNotFound
		}
		return err
	}

	stmt := `
		INSERT INTO teams_players (player_id, team_id, user_id)
		VALUES ($1, $2, $3)`

	args := []any{playerID, teamID, userID}

	result, err := tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		switch {
		case err.Error() == `pq: insert or update on table "teams_players" violates foreign key `+
			`constraint "teams_players_team_id_fkey"`:
			return ErrTeamNotFound
		case err.Error() == `pq: insert or update on table "teams_players" violates foreign key `+
			`constraint "teams_players_user_id_fkey"`:
			return ErrUserNotFound
		case err.Error() == `pq: duplicate key value violates unique constraint `+
			`"teams_players_pkey"`:
			return ErrDuplicatePlayer
		default:
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		return err
	}

	return nil
}

func unassignPlayer(teamID, userID int64, playerPin string, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		DELETE FROM teams_players
		WHERE player_id = (
				SELECT players.id
				FROM players 
				JOIN pins ON players.pin_id = pins.id
				WHERE pins.pin = $1)
			AND team_id = $2 
			AND user_id = $3`

	result, err := tx.ExecContext(ctx, stmt, playerPin, teamID, userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrPlayerNotOnTeam
		default:
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return ErrPlayerNotOnTeam
	}

	return nil
}

func getTeamPlayers(team *Team, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		SELECT pins.id, pins.pin, pins.scope, players.id, players.user_id, players.first_name, 
			players.last_name, players.pref_number, players.is_active, players.created_at, players.version
		FROM teams_players
		JOIN players ON teams_players.player_id = players.id
		JOIN teams ON teams_players.team_id = teams.id
		JOIN pins ON players.pin_id = pins.id
		WHERE teams.user_id = $1 AND teams.id = $2
		ORDER BY players.last_name`

	rows, err := tx.QueryContext(ctx, stmt, team.UserID, team.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil
		default:
			return err
		}
	}
	defer rows.Close()

	players := []*Player{}
	for rows.Next() {
		var player Player
		err := rows.Scan(
			&player.PinId.ID,
			&player.PinId.Pin,
			&player.PinId.Scope,
			&player.ID,
			&player.UserId,
			&player.FirstName,
			&player.LastName,
			&player.PrefNumber,
			&player.IsActive,
			&player.CreatedAt,
			&player.Version,
		)
		if err != nil {
			return err
		}
		players = append(players, &player)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	team.Players = players
	team.Size = len(players)
	return nil
}

func ValidateTeam(v *validator.Validator, team *Team) {
	v.Check(team.Name != "", "name", "must be provided")
	v.Check(len(team.Name) <= 20, "name", "must be 20 characters or less")
}
