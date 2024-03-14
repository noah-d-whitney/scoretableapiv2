package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	"errors"
	"net/http"
)

func (app *application) InsertTeam(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	team := &data.Team{
		Name: input.Name,
	}

	v := validator.New()
	if data.ValidateTeam(v, team); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	team.UserID = app.contextGetUser(r).ID
	team.Size = 0
	pin, err := app.models.Pins.New(data.SCOPE_TEAMS)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	team.PinID = *pin

	err = app.models.Teams.Insert(team)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateTeamName):
			v.AddError("name", "must be unique")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.badRequestResponse(w, r, err)
		}
		err := app.models.Pins.Delete(pin.ID, pin.Scope)
		if err != nil {
			app.logError(r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"team": team}, nil)
}
