package data

import (
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

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

func getGameTeamsPlayers(game *Game, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		SELECT pins.pin, players.id, players.first_name, players.last_name, teams_players.player_number,
			teams_players.lineup_number, (
				SELECT count(*)::int::bool
					FROM teams_players
					JOIN teams ON teams_players.team_id = teams.id
					WHERE player_id = players.id 
						AND lineup_number IS NOT NULL 
						AND teams.is_active IS NOT FALSE)	
			FROM players
				JOIN teams_players ON players.id = teams_players.player_id
				JOIN pins ON players.pin_id = pins.id
			WHERE teams_players.user_id = $1 AND teams_players.team_id = $2
			ORDER BY lineup_number, last_name`

	if game.Teams.Home != nil {
		rows, err := tx.QueryContext(ctx, stmt, game.Teams.Home.UserID, game.Teams.Home.ID)
		if err != nil {
			return err
		}

		players := make([]*Player, 0)
		for rows.Next() {
			var player Player
			err := rows.Scan(
				&player.PinId.Pin,
				&player.ID,
				&player.FirstName,
				&player.LastName,
				&player.Number,
				&player.LineupPos,
				&player.IsActive,
			)
			if err != nil {
				return err
			}
			players = append(players, &player)
		}
		game.Teams.Home.Players = players
	}

	if game.Teams.Away != nil {
		rows, err := tx.QueryContext(ctx, stmt, game.Teams.Away.UserID, game.Teams.Away.ID)
		if err != nil {
			return err
		}

		players := make([]*Player, 0)
		for rows.Next() {
			var player Player
			err := rows.Scan(
				&player.PinId.Pin,
				&player.ID,
				&player.FirstName,
				&player.LastName,
				&player.Number,
				&player.LineupPos,
				&player.IsActive,
			)
			if err != nil {
				return err
			}
			players = append(players, &player)
		}
		game.Teams.Away.Players = players
	}

	return nil
}

func assignGameTeam(gameID, userID int64, teamPin string, teamSide GameTeamSide,
	tx *sql.Tx, ctx context.Context) error {
	getStmt := `
		SELECT teams.id, (SELECT count(*) FROM games_teams WHERE user_id = $1 AND game_id = $2
			AND side = $3)::int::bool, (SELECT count(*) FROM teams_players WHERE user_id = $1 
			AND team_id = teams.id), (SELECT team_size FROM games WHERE user_id = $1 AND id = $2)
		FROM teams
		JOIN pins ON teams.pin_id = pins.id
		WHERE pins.pin = $4 AND teams.user_id = $1`

	var teamID int64
	var assignedTeam bool
	var teamSize int64
	var gameSize int64
	err := tx.QueryRowContext(ctx, getStmt, userID, gameID, teamSide, teamPin).Scan(&teamID,
		&assignedTeam, &teamSize, &gameSize)
	if err != nil {
		return err
	}

	if teamSize < gameSize {
		return ModelValidationErr{Errors: map[string]string{fmt.Sprintf("%s_team_pin",
			teamSide.String()): fmt.Sprintf("team %s has %d players and game requires %d",
			teamPin, teamSize, gameSize)}}
	}

	if assignedTeam {
		err := unassignGameTeam(gameID, userID, teamSide, tx, ctx)
		if err != nil {
			return err
		}
	}

	stmt := `
		INSERT INTO games_teams (user_id, game_id, team_id, side)
		VALUES ($1, $2, $3, $4)`

	args := []any{userID, gameID, teamID, teamSide}

	result, err := tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint `+
			`"games_teams_pkey"`:
			return ModelValidationErr{Errors: map[string]string{
				fmt.Sprintf("%s_team_pin", teamSide): "cannot assign same team",
			}}
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

func unassignGameTeam(gameID, userID int64, teamSide GameTeamSide, tx *sql.Tx,
	ctx context.Context) error {
	stmt := `
		DELETE FROM games_teams
		WHERE user_id = $1 AND game_id = $2 AND side = $3`

	result, err := tx.ExecContext(ctx, stmt, userID, gameID, teamSide)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ModelValidationErr{Errors: map[string]string{
				fmt.Sprintf("%s_team_pin", teamSide): "no team assigned to side",
			}}
		default:
			return err
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return ModelValidationErr{Errors: map[string]string{
			fmt.Sprintf("%s_team_pin", teamSide): "no team assigned to side",
		}}
	}

	return nil
}

func checkTeamConflict(game *Game, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		SELECT pins.pin
		FROM games_teams
		JOIN teams_players ON games_teams.team_id = teams_players.team_id
		JOIN players ON teams_players.player_id = players.id
		JOIN pins ON players.pin_id = pins.id
		WHERE teams_players.team_id = $1 
		INTERSECT SELECT pins.pin
		FROM games_teams
		JOIN teams_players ON games_teams.team_id = teams_players.team_id
		JOIN players ON teams_players.player_id = players.id
		JOIN pins ON players.pin_id = pins.id
		WHERE teams_players.team_id = $2
		`

	if game.Teams.Home == nil || game.Teams.Away == nil {
		return nil
	}

	rows, err := tx.QueryContext(ctx, stmt, game.Teams.Home.ID, game.Teams.Away.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	modelValidationErr := ModelValidationErr{Errors: make(map[string]string)}
	playerPins := make([]string, 0)
	for rows.Next() {
		var playerPin string
		err := rows.Scan(&playerPin)
		if err != nil {
			return err
		}
		playerPins = append(playerPins, playerPin)
	}
	for _, p := range playerPins {
		modelValidationErr.AddError(fmt.Sprintf("player %s", p), "cannot be assigned to both teams in game")
	}
	if !modelValidationErr.Valid() {
		return modelValidationErr
	}

	return nil
}

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
		v.Check(game.PeriodLength != 0, "period_length", "must be provided for timed game")
		v.Check(game.ScoreTarget == nil, "score_target", "cannot be provided for a timed game")
		if !v.Valid() {
			return
		}

		v.Check(game.PeriodLength.Duration() <= 30*time.Minute, "period_count", "must be 30 minutes or less")

		v.Check(*game.PeriodCount > 0, "period_count", "must be greater than 0")
		v.Check(*game.PeriodCount <= 4, "period_count", "must be 4 or less")
	}

	if *game.Type == GameTypeTarget {
		v.Check(game.ScoreTarget != nil, "score_target", "must be provided for target game")
		v.Check(game.PeriodCount == nil, "period_count", "cannot be provided for a target game")
		v.Check(game.PeriodLength == 0, "period_length", "cannot be provided for a target game")
		if !v.Valid() {
			return
		}

		v.Check(*game.ScoreTarget > 0, "score_target", "must be greater than 0")
		v.Check(*game.ScoreTarget <= 100, "score_target", "must be 100 or less")
	}
}
