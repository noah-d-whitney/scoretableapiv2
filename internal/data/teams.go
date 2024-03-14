package data

import (
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
	PinID     Pin       `json:"pin_id"`
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
	stmt := `
		INSERT INTO teams (pin_id, user_id, name, size)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version, is_active`

	args := []any{team.PinID.ID, team.UserID, team.Name, team.Size}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, stmt, args...).Scan(
		&team.ID,
		&team.CreatedAt,
		&team.Version,
		&team.IsActive,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint`+
			` "unq_userid_team_name"`:
			return ErrDuplicateTeamName
		default:
			return err
		}
	}

	return nil
}

func (m *TeamModel) AssignTeamPlayers(team *Team) error {
	stmt := fmt.Sprintf(`
		INSERT INTO teams_players (user_id, team_id, player_id) VALUES %s;`,
		m.GenerateTeamPlayerValues(team))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.db.ExecContext(ctx, stmt)
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

	return nil
}

func (m *TeamModel) GenerateTeamPlayerValues(t *Team) string {
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
