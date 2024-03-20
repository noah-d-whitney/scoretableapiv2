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

var (
	ErrGameNotFound      = errors.New("game could not be found")
	ErrDuplicateGameTeam = errors.New("duplicate game team")
	ErrTeamNotInGame     = errors.New("team not assigned to game")
)

type Game struct {
	ID           int64      `json:"-"`
	UserID       int64      `json:"user_id"`
	PinID        pins.Pin   `json:"pin_id"`
	CreatedAt    time.Time  `json:"-"`
	Version      int64      `json:"-"`
	Status       GameStatus `json:"status"`
	DateTime     *time.Time `json:"date_time"`
	TeamSize     *int64     `json:"team_size"`
	Type         *GameType  `json:"type"`
	PeriodLength *int64     `json:"period_length,omitempty"`
	PeriodCount  *int64     `json:"period_count,omitempty"`
	ScoreTarget  *int64     `json:"score_target,omitempty"`
	HomeTeamPin  string     `json:"home_team_pin,omitempty"`
	AwayTeamPin  string     `json:"away_team_pin,omitempty"`
	Teams        struct {
		Home *Team `json:"home,omitempty"`
		Away *Team `json:"away,omitempty"`
	} `json:"teams,omitempty"`
}

type GameStatus int64

const (
	NOTSTARTED GameStatus = iota
	INPROGRESS
	FINISHED
	CANCELED
)

func (s GameStatus) MarshalJSON() ([]byte, error) {
	switch s {
	case 0:
		return []byte(`"not-started"`), nil
	case 1:
		return []byte(`"in_progress"`), nil
	case 2:
		return []byte(`"finished"`), nil
	case 3:
		return []byte(`"canceled"`), nil
	default:
		return nil, errors.New(`"invalid game status"`)
	}
}

type GameModel struct {
	db *sql.DB
}

type GameType string

const (
	GameTypeTimed  GameType = "timed"
	GameTypeTarget GameType = "target"
)

type GameTeamSide int64

const (
	TeamHome GameTeamSide = iota
	TeamAway
)

func (m *GameModel) Insert(game *Game) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	pin, err := helperModels.Pins.New(pins.PinScopeGames, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	game.PinID = *pin

	stmt := `
		INSERT INTO games (user_id, pin_id, date_time, team_size, 
			period_length, period_count, score_target)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, version, status`

	args := []any{
		game.UserID,
		game.PinID.ID,
		game.DateTime,
		game.TeamSize,
		game.PeriodLength,
		game.PeriodCount,
		game.ScoreTarget,
	}

	err = tx.QueryRowContext(ctx, stmt, args...).Scan(
		&game.ID,
		&game.CreatedAt,
		&game.Version,
		&game.Status,
	)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	if game.HomeTeamPin != "" {
		err := assignGameTeam(game.ID, game.UserID, game.HomeTeamPin, TeamHome, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	if game.AwayTeamPin != "" {
		err := assignGameTeam(game.ID, game.UserID, game.AwayTeamPin, TeamAway, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
	}

	if game.AwayTeamPin != "" || game.HomeTeamPin != "" {
		err := getGameTeams(game, tx, ctx)
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

func getGameTeams(game *Game, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		SELECT pins.id, pins.pin, pins.scope, teams.id, teams.user_id, teams.name, 
			teams.created_at, teams.version, teams.is_active, games_teams.side, (
				SELECT count(*)
				FROM teams_players
				WHERE teams_players.user_id = $1 AND teams_players.team_id = teams.id)	
		FROM games_teams
		JOIN teams ON games_teams.team_id = teams.id
		JOIN games ON games_teams.game_id = games.id
		JOIN pins ON teams.pin_id = pins.id
		WHERE games.user_id = $1 AND games.id = $2
		ORDER BY teams.name`

	rows, err := tx.QueryContext(ctx, stmt, game.UserID, game.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil
		default:
			return err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var team Team
		err := rows.Scan(
			&team.PinID.ID,
			&team.PinID.Pin,
			&team.PinID.Scope,
			&team.ID,
			&team.UserID,
			&team.Name,
			&team.CreatedAt,
			&team.Version,
			&team.IsActive,
			&team.Side,
			&team.Size,
		)
		if err != nil {
			return err
		}
		if team.Side == TeamHome {
			game.Teams.Home = &team
		}
		if team.Side == TeamAway {
			game.Teams.Away = &team
		}
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
}

// TODO check only two teams per game & team sizes are equal
func assignGameTeam(gameID, userID int64, teamPin string, teamSide GameTeamSide, tx *sql.Tx,
	ctx context.Context) error {
	getStmt := `
		SELECT teams.id
		FROM teams
		JOIN pins ON teams.pin_id = pins.id
		WHERE pins.pin = $1 AND teams.user_id = $2`

	var teamID int64
	err := tx.QueryRowContext(ctx, getStmt, teamPin, userID).Scan(&teamID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	stmt := `
		INSERT INTO games_teams (user_id, game_id, team_id, side)
		VALUES ($1, $2, $3, $4)`

	args := []any{userID, gameID, teamID, teamSide}

	result, err := tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		switch {
		case err.Error() == `pq: insert or update on table "games_teams" violates foreign key `+
			`constraint "games_teams_game_id_fkey"`:
			return ErrGameNotFound
		case err.Error() == `pq: insert or update on table "teams_players" violates foreign key `+
			`constraint "games_teams_user_id_fkey"`:
			return ErrUserNotFound
		case err.Error() == `pq: duplicate key value violates unique constraint `+
			`"games_teams_pkey"`:
			return ErrDuplicateGameTeam
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

func unassignGameTeam(gameID, userID int64, teamPin string, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		DELETE FROM games_teams
		WHERE team_id = (
				SELECT teams.id
				FROM teams
				JOIN pins ON teams.pin_id = pins.id
				WHERE pins.pin = $1)
			AND game_id = $2 
			AND user_id = $3`

	result, err := tx.ExecContext(ctx, stmt, teamPin, gameID, userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrTeamNotInGame
		default:
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return ErrTeamNotInGame
	}

	return nil
}

// TODO check if teams are big enough for game

func ValidateGame(v *validator.Validator, game *Game) {
	v.Check(game.DateTime != nil, "date_time", "must be provided")
	v.Check(game.TeamSize != nil, "team_size", "must be provided")
	v.Check(game.Type != nil, "type", "must be provided")
	if game.HomeTeamPin == game.AwayTeamPin {
		v.AddError("home_team_pin", "cannot match away team")
		v.AddError("away_team_pin", "cannot match home team")
	}
	if !v.Valid() {
		return
	}

	v.Check(game.DateTime.After(time.Now()), "date_time", "must be in the future")
	v.Check(*game.TeamSize > 0, "team_size", "must be greater than 0")
	v.Check(*game.TeamSize <= 5, "team_size", "must be 5 or less")
	v.Check(*game.Type == GameTypeTimed || *game.Type == GameTypeTarget, "type",
		fmt.Sprintf(`Must be one of the following: "%s", "%s"`, GameTypeTimed, GameTypeTarget))

	if *game.Type == GameTypeTimed {
		v.Check(game.PeriodCount != nil, "period_count", "must be provided for timed game")
		v.Check(game.PeriodLength != nil, "period_length", "must be provided for timed game")
		v.Check(game.ScoreTarget == nil, "score_target", "cannot be provided for a timed game")
		if !v.Valid() {
			return
		}

		v.Check(*game.PeriodLength > 0, "period_length", "must be greater than 0 seconds")
		v.Check(*game.PeriodLength <= 60*30, "period_count", "must be 20 minutes or less")

		v.Check(*game.PeriodCount > 0, "period_count", "must be greater than 0")
		v.Check(*game.PeriodCount <= 4, "period_count", "must be 4 or less")
	}

	if *game.Type == GameTypeTarget {
		v.Check(game.ScoreTarget != nil, "score_target", "must be provided for target game")
		v.Check(game.PeriodCount == nil, "period_count", "cannot be provided for a target game")
		v.Check(game.PeriodLength == nil, "period_length", "cannot be provided for a target game")
		if !v.Valid() {
			return
		}

		v.Check(*game.ScoreTarget > 0, "score_target", "must be greater than 0")
		v.Check(*game.ScoreTarget <= 100, "score_target", "must be 100 or less")
	}
}
