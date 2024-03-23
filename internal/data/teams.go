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
	ErrDuplicateTeamName = NewModelValidationErr("name", "must be unique")
	ErrPlayerNotFound    = NewModelValidationErr("player_ids",
		"one or more player could not be found")
	ErrDuplicatePlayer = NewModelValidationErr("player_ids",
		"cannot assign player to same team more than once")
	ErrPlayerNotOnTeam = NewModelValidationErr("player_ids", "cannot find player on team")
	ErrPlayerNumNotUnq = NewModelValidationErr("player_ids", "duplicate player number on team")
)

type Team struct {
	ID           int64          `json:"-"`
	PinID        pins.Pin       `json:"pin"`
	UserID       int64          `json:"-"`
	Name         string         `json:"name"`
	Size         int            `json:"size"`
	CreatedAt    time.Time      `json:"-"`
	Version      int32          `json:"-"`
	IsActive     bool           `json:"is_active"`
	PlayerIDs    []string       `json:"-"`
	Players      []*Player      `json:"players,omitempty"`
	PlayerNums   map[string]int `json:"-"`
	PlayerLineup []string       `json:"-"`
	Side         GameTeamSide   `json:"-"`
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

	if len(team.PlayerIDs) == 0 && len(team.PlayerLineup) != 0 {
		return NewModelValidationErr("player_lineup", "no players to assign to lineup")
	}

	if len(team.PlayerIDs) != 0 {
		for _, p := range team.PlayerIDs {
			err := assignPlayer(team, p, tx, ctx)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return rollbackErr
				}
				return err
			}
		}

		if len(team.PlayerLineup) != 0 {
			err := assignTeamLineup(team, tx, ctx)
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
	Metadata, error) {
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
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint`+
			`"unq_userid_team_name"`:
			return ErrDuplicateTeamName
		default:
			return err
		}
	}

	err = tx.QueryRowContext(ctx, stmt, args...).Scan(&team.Version)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	// ASSIGN AND UNASSIGN PLAYERS
	assign, unassign := parsePinList(team.PlayerIDs)

	for _, pin := range assign {
		err := assignPlayer(team, pin, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}

		err = checkPlayerConflict(team.UserID, pin, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	for _, pin := range unassign {
		err := unassignPlayer(team, pin, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	team.PlayerIDs = nil

	if len(team.PlayerNums) != 0 {
		err := editPlayerNumbers(team, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	team.PlayerNums = nil

	if len(team.PlayerLineup) != 0 {
		err := assignTeamLineup(team, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	team.PlayerLineup = nil

	err = getTeamPlayers(team, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	checkSizeStmt := `
		SELECT pins.pin, games.team_size
		FROM games
		JOIN games_teams ON games.id = games_teams.game_id
		JOIN pins ON games.pin_id = pins.id
		WHERE games_teams.user_id = $1 AND games_teams.team_id = $2 AND games.team_size > $3`

	rows, err := tx.QueryContext(ctx, checkSizeStmt, team.UserID, team.ID, team.Size)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			break
		default:
			return err
		}
	}
	defer rows.Close()

	valErr := ModelValidationErr{Errors: make(map[string]string)}
	for rows.Next() {
		var gamePin string
		var gameSize int64
		err := rows.Scan(&gamePin, &gameSize)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
		valErr.AddError(fmt.Sprintf("game %s", gamePin),
			fmt.Sprintf("not enough players (%d) for game (%d needed)", team.Size, gameSize))
	}
	if !valErr.Valid() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return valErr
	}

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

func editPlayerNumbers(team *Team, tx *sql.Tx, ctx context.Context) error {
	getPinsStmt := `
		SELECT pins.pin
			FROM pins
				JOIN players ON pins.id = players.pin_id
				JOIN teams_players ON players.id = teams_players.player_id
			WHERE teams_players.user_id = $1 AND teams_players.team_id = $2`

	rows, err := tx.QueryContext(ctx, getPinsStmt, team.UserID, team.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	playerPins := make([]string, 0)
	for rows.Next() {
		var pin string
		err := rows.Scan(&pin)
		if err != nil {
			return err
		}
		playerPins = append(playerPins, pin)
	}

	assignNumStmt := `
		UPDATE teams_players
			SET player_number = $1
			WHERE user_id = $2 AND team_id = $3 AND player_id = (
				SELECT players.id
					FROM players 
						JOIN pins ON players.pin_id = pins.id
					WHERE pins.pin = $4)`

	for _, p := range playerPins {
		if num, exists := team.PlayerNums[p]; exists {
			if num < 1 {
				return NewModelValidationErr("player_nums", fmt.Sprintf(`player's (%s) number (`+
					`%d) must be greater than 0`, p, num))
			}
			if num > 99 {
				return NewModelValidationErr("player_nums", fmt.Sprintf(`player's (%s) number (`+
					`%d) must be less than 100`, p, num))
			}
			args := []any{num, team.UserID, team.ID, p}
			result, err := tx.ExecContext(ctx, assignNumStmt, args...)
			if err != nil {
				switch {
				case err.Error() == `pq: duplicate key value violates unique constraint `+
					`"teams_players_player_number_unq"`:
					e := ErrPlayerNumNotUnq
					e.Errors["player_ids"] = fmt.Sprintf(`player's (%s) dersired number (%d) is already `+
						`in use on this team.`, p, num)
					return e
				default:
					return err
				}
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil || rowsAffected != 1 {
				return err
			}

			delete(team.PlayerNums, p)
		}
	}

	return nil
}

func assignPlayer(team *Team, playerPin string, tx *sql.Tx, ctx context.Context) error {
	getStmt := `
		SELECT players.id, players.pref_number
		FROM players
		JOIN public.pins ON pins.id = players.pin_id
		WHERE pins.pin = $1 AND players.user_id = $2`

	var playerID string
	var number int

	err := tx.QueryRowContext(ctx, getStmt, playerPin, team.UserID).Scan(&playerID, &number)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrPlayerNotFound
		}
		return err
	}

	if num, exists := team.PlayerNums[playerPin]; exists {
		number = num
	}

	stmt := `
		INSERT INTO teams_players (player_id, team_id, user_id, player_number, lineup_number)
		VALUES ($1, $2, $3, $4, $5)`

	args := []any{playerID, team.ID, team.UserID, number, nil}
	println(playerID)

	result, err := tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint `+
			`"teams_players_pkey"`:
			e := ErrDuplicatePlayer
			e.Errors["player_ids"] = fmt.Sprintf("cannot assign player %s to team more than once",
				playerPin)
			return e
		case err.Error() == `pq: duplicate key value violates unique constraint `+
			`"teams_players_player_number_unq"`:
			e := ErrPlayerNumNotUnq
			e.Errors["player_ids"] = fmt.Sprintf(`player's (%s) pref number (%d) is already `+
				`in use on this team.`, playerPin, number)
			return e
		default:
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		return err
	}

	team.Size++

	return nil
}

func unassignPlayer(team *Team, playerPin string, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		DELETE FROM teams_players
		WHERE player_id = (
				SELECT players.id
				FROM players 
				JOIN pins ON players.pin_id = pins.id
				WHERE pins.pin = $1)
			AND team_id = $2 
			AND user_id = $3`

	result, err := tx.ExecContext(ctx, stmt, playerPin, team.ID, team.UserID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			e := ErrPlayerNotOnTeam
			e.Errors["player_ids"] = fmt.Sprintf("player %s is not on team", playerPin)
			return e
		default:
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		e := ErrPlayerNotOnTeam
		e.Errors["player_ids"] = fmt.Sprintf("player %s is not on team", playerPin)
		return e
	}

	team.Size--

	return nil
}

func assignTeamLineup(team *Team, tx *sql.Tx, ctx context.Context) error {
	resetLineupStmt := `
		UPDATE teams_players
		SET lineup_number = null
		WHERE user_id = $1 AND team_id = $2`

	_, err := tx.ExecContext(ctx, resetLineupStmt, team.UserID, team.ID)
	if err != nil {
		return err
	}

	stmt := `
		UPDATE teams_players
		SET lineup_number = $1
		WHERE user_id = $2 AND team_id = $3 AND player_id = (
			SELECT players.id
			FROM players
			JOIN pins ON pins.id = players.pin_id
			WHERE pins.pin = $4)`

	for i, p := range team.PlayerLineup {
		args := []any{i + 1, team.UserID, team.ID, p}
		result, err := tx.ExecContext(ctx, stmt, args...)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return NewModelValidationErr("player_lineup", fmt.Sprintf(
					"player (%s) at lineup #%d could not be found on team", p, i))
			default:
				return err
			}
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected != 1 {
			return NewModelValidationErr("player_lineup", fmt.Sprintf(
				"player (%s) at lineup #%d could not be found on team", p, i))
		}
	}

	team.PlayerLineup = nil
	return nil
}

func getTeamPlayers(team *Team, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		SELECT pins.id, pins.pin, pins.scope, players.id, players.user_id, players.first_name, 
			players.last_name, players.created_at, players.version, 
			teams_players.player_number, teams_players.lineup_number, (
				SELECT count(*)::int::bool
					FROM teams_players
					WHERE player_id = players.id AND lineup_number IS NOT NULL)	
			FROM teams_players
				JOIN players ON teams_players.player_id = players.id
				JOIN teams ON teams_players.team_id = teams.id
				JOIN pins ON players.pin_id = pins.id
				WHERE teams.user_id = $1 AND teams.id = $2
				ORDER BY teams_players.lineup_number, players.last_name`

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
			&player.CreatedAt,
			&player.Version,
			&player.Number,
			&player.LineupPos,
			&player.IsActive,
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

func checkPlayerConflict(userID int64, playerPin string, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		SELECT pins.pin
			FROM games_teams
        		JOIN teams_players ON games_teams.team_id = teams_players.team_id
        		JOIN games ON games_teams.game_id = games.id
        		JOIN pins ON games.pin_id = pins.id
			WHERE games_teams.user_id = $1
  				AND teams_players.player_id = (
					SELECT players.id
						FROM players
							JOIN pins ON players.pin_id = pins.id
							WHERE pins.pin = $2)		
    			AND games_teams.side = 0
		INTERSECT SELECT pins.pin
			FROM games_teams
    			JOIN teams_players ON games_teams.team_id = teams_players.team_id
    			JOIN games ON games_teams.game_id = games.id
    			JOIN pins ON games.pin_id = pins.id
			WHERE games_teams.user_id = $1
  				AND teams_players.player_id = (
					SELECT players.id
						FROM players
							JOIN pins ON players.pin_id = pins.id
							WHERE pins.pin = $2)		
  				AND games_teams.side = 1`

	rows, err := tx.QueryContext(ctx, stmt, userID, playerPin)
	if err != nil {
		return err
	}
	defer rows.Close()

	modelValidationErr := ModelValidationErr{Errors: make(map[string]string)}
	gamePins := make([]string, 0)
	for rows.Next() {
		var gamePin string
		err := rows.Scan(&gamePin)
		if err != nil {
			return err
		}
		gamePins = append(gamePins, gamePin)
	}

	for _, p := range gamePins {
		modelValidationErr.AddError(fmt.Sprintf("game %s", p),
			"player cannot be assigned to both teams in game")
	}
	if !modelValidationErr.Valid() {
		return modelValidationErr
	}

	return nil
}

func ValidateTeam(v *validator.Validator, team *Team) {
	v.Check(team.Name != "", "name", "must be provided")
	v.Check(len(team.Name) <= 20, "name", "must be 20 characters or less")
}
