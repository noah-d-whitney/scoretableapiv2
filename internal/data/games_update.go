package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (m *GameModel) Update(game *Game) error {
	stmt := `
		UPDATE games
			SET date_time = $1, team_size = $2, period_length = $3, period_count = $4,
				score_target = $5
			WHERE user_id = $6
			  	AND id = $7
				AND version = $8
			RETURNING version`

	args := []any{game.DateTime, game.TeamSize, game.PeriodLength, game.PeriodCount, game.ScoreTarget,
		game.UserID, game.ID, game.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx, stmt, args...).Scan(&game.Version)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	if game.HomeTeamPin != nil {
		if *game.HomeTeamPin == "-" {
			err := unassignGameTeam(game.ID, game.UserID, TeamHome, tx, ctx)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return rollbackErr
				}
				return err
			}
			game.Teams.Home = nil
		} else {
			err := assignGameTeam(game.ID, game.UserID, *game.HomeTeamPin, TeamHome, tx, ctx)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return rollbackErr
				}
				return err
			}
		}
	}

	if game.AwayTeamPin != nil {
		if *game.AwayTeamPin == "-" {
			err := unassignGameTeam(game.ID, game.UserID, TeamAway, tx, ctx)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return rollbackErr
				}
				return err
			}
			game.Teams.Away = nil
		} else {
			err := assignGameTeam(game.ID, game.UserID, *game.AwayTeamPin, TeamAway, tx, ctx)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return rollbackErr
				}
				return err
			}
		}
	}

	err = validateGameSize(game, tx, ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	if game.AwayTeamPin != nil || game.HomeTeamPin != nil {
		err := getGameTeams(game, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return err
		}
		game.AwayTeamPin = nil
		game.HomeTeamPin = nil

		err = checkTeamConflict(game, tx, ctx)
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

func validateGameSize(game *Game, tx *sql.Tx, ctx context.Context) error {
	stmt := `
		SELECT LEAST((
    		SELECT count(*)
        		FROM teams_players tp
					JOIN public.games_teams gt on tp.team_id = gt.team_id
        	WHERE gt.game_id = g.id AND side = 0
    	)::bigint, (
    		SELECT count(*)
        		FROM teams_players tp
                 	JOIN public.games_teams gt on tp.team_id = gt.team_id
        		WHERE gt.game_id = g.id AND side = 1
    	)::bigint) as game_team_size_min
		FROM games g
		WHERE g.user_id = $1 AND g.id = $2`

	var maxTeamSize int64
	err := tx.QueryRowContext(ctx, stmt, game.UserID, game.ID).Scan(&maxTeamSize)
	if err != nil {
		return err
	}

	// maxTeamSize is 0 if game doesn't have both teams assigned
	if maxTeamSize == 0 {
		return nil
	}

	if game.TeamSize > maxTeamSize {
		return ModelValidationErr{Errors: map[string]string{
			"team_size": fmt.Sprintf(`specified team size (`+
				`%d) must be at most the size of smallest team (%d)`, game.TeamSize, maxTeamSize),
		}}
	}

	return nil
}
