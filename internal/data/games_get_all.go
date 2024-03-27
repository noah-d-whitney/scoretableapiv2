package data

import (
	"ScoreTableApi/internal/validator"
	"context"
	"fmt"
	"github.com/lib/pq"
	"slices"
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
	Pag        Metadata `json:"pag"`
	*DateRange `json:"date_range,omitempty"`
	TeamPins   []string     `json:"team_pins,omitempty"`
	PlayerPins []string     `json:"player_pins,omitempty"`
	Type       GameType     `json:"type,omitempty"`
	TeamSize   []int64      `json:"team_size,omitempty"`
	Status     []GameStatus `json:"status,omitempty"`
	Includes   []string     `json:"includes,omitempty"`
}

// todo get sorted games for status

func (m *GameModel) GetAll(userID int64, filters GamesFilter, includes []string) ([]*Game,
	GamesMetadata, error) {
	stmt := fmt.Sprintf(`
		SELECT count(*) OVER(), pin_id, pin, scope, id, user_id, created_at, version, status, date_time, 
			team_size, period_length, period_count, score_target, type
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
			ORDER BY %s %s, id ASC
			LIMIT $16 OFFSET $17`, filters.Filters.sortColumn(), filters.Filters.sortDirection())
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
		filters.DateRange.AfterDate != nil,
		filters.DateRange.AfterDate,
		filters.DateRange.BeforeDate != nil,
		filters.DateRange.BeforeDate,
		filters.PlayerPins != nil,
		pq.Array(filters.PlayerPins),
		filters.Type != "",
		filters.Type,
		filters.TeamSize != nil,
		pq.Array(filters.TeamSize),
		filters.Status != nil,
		pq.Array(filters.Status),
		filters.Filters.limit(),
		filters.Filters.offset(),
	}

	rows, err := tx.QueryContext(ctx, stmt, args...)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, GamesMetadata{}, rollbackErr
		}
		return nil, GamesMetadata{}, err
	}

	totalRecords := 0
	games := make([]*Game, 0)
	for rows.Next() {
		var game Game
		err := rows.Scan(
			&totalRecords,
			&game.PinID.ID,
			&game.PinID.Pin,
			&game.PinID.Scope,
			&game.ID,
			&game.UserID,
			&game.CreatedAt,
			&game.Version,
			&game.Status,
			&game.DateTime,
			&game.TeamSize,
			&game.PeriodLength,
			&game.PeriodCount,
			&game.ScoreTarget,
			&game.Type,
		)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, GamesMetadata{}, rollbackErr
			}
			return nil, GamesMetadata{}, err
		}
		games = append(games, &game)
	}

	for _, g := range games {
		err := getGameTeams(g, tx, ctx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, GamesMetadata{}, rollbackErr
			}
			return nil, GamesMetadata{}, err
		}
		g.HomeTeamPin = nil
		g.AwayTeamPin = nil
	}

	if slices.Contains(includes, "players") {
		for _, g := range games {
			err := getGameTeamsPlayers(g, tx, ctx)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					return nil, GamesMetadata{}, rollbackErr
				}
				return nil, GamesMetadata{}, err
			}
		}
	}

	metadata := calculateGamesMetadata(totalRecords, filters, includes)

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
		Includes:   includes,
	}

	if !f.DateRange.IsEmpty() {
		metadata.DateRange = &f.DateRange
	}

	return metadata
}

// TODO add validation for all fields
// TODO Refactor get all games metadata and validation

func ValidateGamesFilter(v *validator.Validator, f GamesFilter) {
	ValidateFilters(v, f.Filters)
	if f.DateRange.AfterDate != nil {
		v.Check(f.DateRange.AfterDate.After(GAME_MIN_DATE), "after_date", "must be in 2024 or after")
	}
	if f.DateRange.BeforeDate != nil {
		v.Check(f.DateRange.BeforeDate.After(GAME_MIN_DATE), "before_date", "must be in 2024 or after")
	}
	if f.DateRange.IsFull() {
		v.Check(f.DateRange.BeforeDate.After(*f.DateRange.AfterDate), "start_date",
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
