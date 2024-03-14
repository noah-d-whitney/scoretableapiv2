package data

import (
	"context"
	"database/sql"
	"time"
)

type Team struct {
	ID        int64
	PinID     string
	UserID    int64
	Name      string
	Size      int
	CreatedAt time.Time
	Version   int32
	IsActive  bool
}

type TeamModel struct {
	db *sql.DB
}

func (m *TeamModel) Insert(team *Team) error {
	stmt := `
		INSERT INTO teams (pin_id, user_id, name, size)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version, is_active`

	args := []any{team.PinID, team.UserID, team.Name, team.Size}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRowContext(ctx, stmt, args...).Scan(
		&team.ID,
		&team.CreatedAt,
		&team.Version,
		&team.IsActive,
	)
	if err != nil {
		return err
	}

	return nil
}
