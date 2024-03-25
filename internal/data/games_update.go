package data

import (
	"context"
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

}
