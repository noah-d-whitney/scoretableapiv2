package data

import (
	"ScoreTableApi/internal/validator"
	"context"
	"github.com/lib/pq"
	"time"
)

//TODO add "valid/startable" query oto game to signignify if game is valid to start
//, teamPins []string, dateRange DateRange,
//includes []string, filters Filters

var (
	GAME_MIN_DATE = time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
)

type GamesFilter struct {
	Filters
	DateRange
	TeamPins   []string
	PlayerPins []string
	Type       GameType
	TeamSize   []int64
	Status     []GameStatus
}

type GamesMetadata struct {
	Metadata
	StartDate  *time.Time   `json:"start_date,omitempty"`
	EndDate    *time.Time   `json:"end_date,omitempty"`
	TeamPins   []string     `json:"team_pins,omitempty"`
	PlayerPins []string     `json:"player_pins,omitempty"`
	Type       GameType     `json:"type,omitempty"`
	TeamSize   []int64      `json:"team_size,omitempty"`
	Status     []GameStatus `json:"status,omitempty"`
}

// todo get sorted games for status

func (m *GameModel) GetAll(userID int64, filters GamesFilter) ([]*Game, GamesMetadata,
	error) {
	stmt := `
		SELECT pin 
			FROM games_view
			WHERE games_view.user_id = $1
			AND (($2 IS FALSE)
				OR games_view.home_team_pin = ANY($3) 
				OR games_view.away_team_pin = ANY($3))
			AND (($4 IS FALSE)
				OR games_view.date_time > $5)
			AND (($6 IS FALSE)
				OR games_view.date_time <= $7)	
			AND (($8 IS FALSE)
			    OR games_view.player_pins @> $9::text[])
			AND (($10 IS FALSE)
				OR games_view.type = $11)
			AND (($12 IS FALSE)
				OR games_view.team_size = ANY($13::integer[]))
			AND (($14 IS FALSE)
				OR games_view.status = ANY($15::integer[]))
			`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, GamesMetadata{}, err
	}

	args := []any{
		userID,
		filters.TeamPins != nil,
		pq.Array(filters.TeamPins),
		!filters.DateRange.Start.IsZero(),
		filters.DateRange.Start,
		!filters.DateRange.End.IsZero(),
		filters.DateRange.End,
		filters.PlayerPins != nil,
		pq.Array(filters.PlayerPins),
		filters.Type != "",
		filters.Type,
		filters.TeamSize != nil,
		pq.Array(filters.TeamSize),
		filters.Status != nil,
		pq.Array(filters.Status),
	}

	rows, err := tx.QueryContext(ctx, stmt, args...)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, GamesMetadata{}, rollbackErr
		}
		return nil, GamesMetadata{}, err
	}

	games := make([]*Game, 0)
	for rows.Next() {
		var game Game
		err := rows.Scan(&game.PinID.Pin)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, GamesMetadata{}, rollbackErr
			}
			return nil, GamesMetadata{}, err
		}
		games = append(games, &game)
	}

	metadata := calculateGamesMetadata(5, filters)

	return games, metadata, nil
}

func calculateGamesMetadata(totalRecords int, f GamesFilter) GamesMetadata {
	if totalRecords == 0 {
		return GamesMetadata{}
	}

	var startDate *time.Time
	var endDate *time.Time
	if f.DateRange.Start.IsZero() {
		startDate = nil
	} else {
		startDate = &f.DateRange.Start
	}
	if f.DateRange.End.IsZero() {
		startDate = nil
	} else {
		startDate = &f.DateRange.End
	}

	return GamesMetadata{
		Metadata:   calculateMetadata(totalRecords, f.Filters.Page, f.Filters.PageSize),
		StartDate:  startDate,
		EndDate:    endDate,
		TeamPins:   f.TeamPins,
		PlayerPins: f.PlayerPins,
		Type:       f.Type,
		TeamSize:   f.TeamSize,
		Status:     f.Status,
	}
}

func ValidateGamesFilter(v *validator.Validator, f GamesFilter) {
	ValidateFilters(v, f.Filters)
	if !f.DateRange.Start.IsZero() {
		v.Check(f.DateRange.Start.After(GAME_MIN_DATE), "start_date", "must be in 2024 or after")
	}
	if !f.DateRange.End.IsZero() {
		v.Check(f.DateRange.End.After(GAME_MIN_DATE), "end_date", "must be in 2024 or after")
	}
	if !f.DateRange.Start.IsZero() && !f.DateRange.End.IsZero() {
		v.Check(f.DateRange.Start.After(f.DateRange.End), "start_date", "cannot be after end date")
	}
	if f.Type != "" {
		v.Check(f.Type == GameTypeTimed || f.Type == GameTypeTarget, "type",
			`must be either "timed" or "target"`)
	}
	if f.TeamSize != nil {
		v.Check(len(f.TeamSize) < 5, "team_size", "must not contain more than 5 selections")
		for _, i := range f.TeamSize {
			v.Check(i <= 5 && i > 0, "team_size", "must be an integer 1-5")
		}
	}
}
