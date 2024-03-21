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
		DateTime     *time.Time `json:"date_time"`
		TeamSize     *int64     `json:"team_size"`
		Type         *string    `json:"type"`
		PeriodLength *int64     `json:"period_length"`
		PeriodCount  *int64     `json:"period_count"`
		ScoreTarget  *int64     `json:"score_target"`
		HomeTeamPin  string     `json:"home_team_pin"`
		AwayTeamPin  string     `json:"away_team_pin"`
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
		PeriodLength: (*data.PeriodLength)(input.PeriodLength),
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
