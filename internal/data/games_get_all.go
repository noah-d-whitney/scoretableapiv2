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
	Filters    `json:"filters"`
	DateRange  `json:"date_range"`
	TeamPins   []string     `json:"team_pins,omitempty"`
	PlayerPins []string     `json:"player_pins,omitempty"`
	Type       GameType     `json:"type,omitempty"`
	TeamSize   []int64      `json:"team_size,omitempty"`
	Status     []GameStatus `json:"status,omitempty"`
}

type GamesMetadata struct {
	Pag     Metadata    `json:"pag"`
	Filters GamesFilter `json:"filters"`
	//*DateRange `json:"date_range"`
	//TeamPins   []string     `json:"team_pins,omitempty"`
	//PlayerPins []string     `json:"player_pins,omitempty"`
	//Type       GameType     `json:"type,omitempty"`
	//TeamSize   []int64      `json:"team_size,omitempty"`
	//Status     []GameStatus `json:"status,omitempty"`
}

// todo get sorted games for status

func (m *GameModel) GetAll(userID int64, filters GamesFilter, includes []string) ([]*Game,
	GamesMetadata, error) {
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
		!filters.DateRange.AfterDate.IsZero(),
		filters.DateRange.AfterDate,
		!filters.DateRange.BeforeDate.IsZero(),
		filters.DateRange.BeforeDate,
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

	metadata := calculateGamesMetadata(5, filters, includes)

	return games, metadata, nil
}

//TODO move games metadata to game_types
//TODO scan all fields from game, implement includes feature

func calculateGamesMetadata(totalRecords int, f GamesFilter, includes []string) GamesMetadata {
	if totalRecords == 0 {
		return GamesMetadata{}
	}

	metadata := GamesMetadata{
		Pag:        calculateMetadata(totalRecords, f.Filters.Page, f.Filters.PageSize),
		PlayerPins: f.PlayerPins,
		Type:       f.Type,
		TeamSize:   f.TeamSize,
		Status:     f.Status,
	}

	if !f.DateRange.IsZero() {
		metadata.DateRange = &f.DateRange
	}

	return metadata
}

// TODO add validation for all fields

func ValidateGamesFilter(v *validator.Validator, f GamesFilter) {
	ValidateFilters(v, f.Filters)
	if !f.DateRange.AfterDate.IsZero() {
		v.Check(f.DateRange.AfterDate.After(GAME_MIN_DATE), "after_date", "must be in 2024 or after")
	}
	if !f.DateRange.BeforeDate.IsZero() {
		v.Check(f.DateRange.BeforeDate.After(GAME_MIN_DATE), "before_date", "must be in 2024 or after")
	}
	if !f.DateRange.AfterDate.IsZero() && !f.DateRange.BeforeDate.IsZero() {
		v.Check(f.DateRange.AfterDate.After(f.DateRange.BeforeDate), "start_date",
			"cannot be after end date")
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
