package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	"fmt"
	"net/http"
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
	}

	if data.ValidateGame(v, game); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Games.Insert(game)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/game/%s", game.PinID.Pin))
	err = app.writeJSON(w, http.StatusCreated, envelope{"game": game}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
