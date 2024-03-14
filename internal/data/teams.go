package data

import (
	"ScoreTableApi/internal/validator"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDuplicateTeamName = errors.New("duplicate team name")
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

func ValidateTeam(v *validator.Validator, team *Team) {
	v.Check(team.Name != "", "name", "must be provided")
	v.Check(len(team.Name) <= 20, "name", "must be 20 characters or less")
}
