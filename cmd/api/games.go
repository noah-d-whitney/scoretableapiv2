package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
	"time"
)

func (app *application) InsertGame(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DateTime     *time.Time        `json:"date_time"`
		TeamSize     *int64            `json:"team_size"`
		Type         *string           `json:"type"`
		PeriodLength data.PeriodLength `json:"period_length"`
		PeriodCount  *int64            `json:"period_count"`
		ScoreTarget  *int64            `json:"score_target"`
		HomeTeamPin  string            `json:"home_team_pin"`
		AwayTeamPin  string            `json:"away_team_pin"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userID := app.contextGetUser(r).ID
	v := validator.New()

	game := &data.Game{
		UserID:       userID,
		DateTime:     input.DateTime,
		TeamSize:     input.TeamSize,
		Type:         (*data.GameType)(input.Type),
		PeriodLength: input.PeriodLength,
		PeriodCount:  input.PeriodCount,
		ScoreTarget:  input.ScoreTarget,
		HomeTeamPin:  input.HomeTeamPin,
		AwayTeamPin:  input.AwayTeamPin,
	}

	if data.ValidateGame(v, game); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Games.Insert(game)
	if err != nil {
		var modelValidationErr data.ModelValidationErr
		switch {
		case errors.As(err, &modelValidationErr):
			app.failedValidationResponse(w, r, modelValidationErr.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/game/%s", game.PinID.Pin))
	err = app.writeJSON(w, http.StatusCreated, envelope{"game": game}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetGame(w http.ResponseWriter, r *http.Request) {
	userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	game, err := app.models.Games.Get(userID, pin)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"game": game}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return
}

func (app *application) GetAllGames(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	v := validator.New()
	userID := app.contextGetUser(r).ID
	filters := data.GamesFilter{}

	filters.DateRange.Start = app.readDate(qs, "start_date", time.Time{}, v)
	filters.DateRange.End = app.readDate(qs, "end_date", time.Time{}, v)
	if !filters.DateRange.End.IsZero() {
		filters.DateRange.End = filters.DateRange.End.Add(24 * time.Hour)
	}
	filters.TeamPins = app.readCSV(qs, "team_pins", nil)
	filters.PlayerPins = app.readCSV(qs, "player_pins", nil)
	filters.Type = data.GameType(app.readString(qs, "type", ""))
	filters.TeamSize = app.readCSInt(qs, "team_size", nil, v)
	filters.Status = app.readCSGameStatus(qs, nil, v)

	filters.Filters.Page = app.readInt(qs, "page", 1, v)
	filters.Filters.PageSize = app.readInt(qs, "page_size", 5, v)
	filters.Filters.Sort = app.readString(qs, "sort", "name")
	filters.Filters.SortSafeList = []string{"pin", "name", "-pin", "-name"}

	if data.ValidateGamesFilter(v, filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	games, metadata, err := app.models.Games.GetAll(userID, filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "games": games}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return
}

func (app *application) DeleteGame(w http.ResponseWriter, r *http.Request) {
	userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	err := app.models.Games.Delete(userID, pin)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"message": fmt.Sprintf("game (%s) successfully deleted", pin)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return
}
