package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func (app *application) InsertTeam(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string  `json:"name"`
		PlayerIDs []int64 `json:"player_ids"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	team := &data.Team{
		Name:      input.Name,
		PlayerIDs: input.PlayerIDs,
	}

	v := validator.New()
	if data.ValidateTeam(v, team); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	team.UserID = app.contextGetUser(r).ID

	err = app.models.Teams.Insert(team)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateTeamName):
			v.AddError("name", "must be unique")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrDuplicatePlayer):
			v.AddError("player_ids", "must not have duplicate player ids")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrPlayerNotFound):
			v.AddError("player_ids", "one or more player could not be found")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.badRequestResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/team/%s", team.PinID.Pin))
	err = app.writeJSON(w, http.StatusCreated, envelope{"team": team}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		app.notFoundResponse(w, r)
		return
	}
	userId := app.contextGetUser(r).ID

	err = app.models.Teams.Delete(id, userId)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "team successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
