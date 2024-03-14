package data

import (
	"ScoreTableApi/internal/pins"
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrDuplicateTeamName = errors.New("duplicate team name")
	ErrPlayerNotFound    = errors.New("player(s) not found")
	ErrTeamNotFound      = errors.New("team not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicatePlayer   = errors.New("duplicate player team assignment")
)

type Team struct {
	ID        int64     `json:"id"`
	PinID     pins.Pin  `json:"pin_id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"-"`
	IsActive  bool      `json:"is_active"`
	PlayerIDs []int64   `json:"-"`
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

	pin, err := helperModels.Pins.New(pins.PinScopeTeams, tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	team.PinID = *pin

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
		err := assignPlayers(team, tx, ctx)
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

func (m *TeamModel) Delete(teamID, userID int64) error {
	stmt := `
		DELETE FROM teams
		WHERE id = $1 AND user_id = $2
		RETURNING pin_id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var pinID int64
	err = tx.QueryRowContext(ctx, stmt, teamID, userID).Scan(&pinID)
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

func assignPlayers(team *Team, tx *sql.Tx, ctx context.Context) error {
	insertStmt := fmt.Sprintf(`
		INSERT INTO teams_players (user_id, team_id, player_id) 
			VALUES %s;`, generateTeamPlayerValues(team))
	adjSizeStmt := `
		UPDATE teams
			SET size = size + $1, version = version + 1
			WHERE id = $2 AND size = $3 AND version = $4
			RETURNING size, version`

	args := []any{len(team.PlayerIDs), team.ID, team.Size, team.Version}

	_, err := tx.ExecContext(ctx, insertStmt)
	if err != nil {
		switch {
		case err.Error() == `pq: insert or update on table "teams_players" violates foreign key `+
			`constraint "teams_players_team_id_fkey"`:
			return ErrTeamNotFound
		case err.Error() == `pq: insert or update on table "teams_players" violates foreign key `+
			`constraint "teams_players_player_id_fkey"`:
			return ErrPlayerNotFound
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

	err = tx.QueryRowContext(ctx, adjSizeStmt, args...).Scan(&team.Size, &team.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		}
		return err
	}

	return nil
}

func generateTeamPlayerValues(t *Team) string {
	var output []string
	for _, pid := range t.PlayerIDs {
		value := fmt.Sprintf("(%d, %d, %d)", t.UserID, t.ID, pid)
		output = append(output, value)
	}
	return strings.Join(output, ", ")
}

func ValidateTeam(v *validator.Validator, team *Team) {
	v.Check(team.Name != "", "name", "must be provided")
	v.Check(len(team.Name) <= 20, "name", "must be 20 characters or less")
}
