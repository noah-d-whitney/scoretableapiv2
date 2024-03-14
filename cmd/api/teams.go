package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	"errors"
	"net/http"
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

	if len(team.PlayerIDs) == 0 {
		err = app.writeJSON(w, http.StatusCreated, envelope{"team": team}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Teams.AssignTeamPlayers(team)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPlayerNotFound):
			v.AddError("player_ids", "one or more players cannot be found")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrDuplicatePlayer):
			v.AddError("player_ids", "duplicate player specified for same team")
			app.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, data.ErrTeamNotFound) || errors.Is(err, data.ErrUserNotFound):
			app.logger.PrintFatal(err, nil)
			app.serverErrorResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

}

func (app *application) GetTeamInsert(w http.ResponseWriter, r *http.Request) {
	err := app.models.Teams.AssignTeamPlayers(&data.Team{
		ID:     1,
		UserID: 7,
	}, []int64{1, 3})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"rows": "ok"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
