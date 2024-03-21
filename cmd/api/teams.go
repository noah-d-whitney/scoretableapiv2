package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"slices"
	"strings"
)

func (app *application) InsertTeam(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string         `json:"name"`
		PlayerIDs  []string       `json:"player_ids"`
		PlayerNums map[string]int `json:"player_nums"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	team := &data.Team{
		Name:       input.Name,
		PlayerIDs:  input.PlayerIDs,
		PlayerNums: input.PlayerNums,
	}

	v := validator.New()
	if data.ValidateTeam(v, team); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	team.UserID = app.contextGetUser(r).ID

	err = app.models.Teams.Insert(team)
	if err != nil {
		var modelValidationErr data.ModelValidationErr
		switch {
		case errors.As(err, &modelValidationErr):
			app.failedValidationResponse(w, r, modelValidationErr.Errors)
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
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

func (app *application) GetTeam(w http.ResponseWriter, r *http.Request) {
	userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	team, err := app.models.Teams.Get(userID, pin)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"team": team}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetAllTeams(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string
		Includes struct {
			Values   []string
			SafeList []string
		}
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()
	userID := app.contextGetUser(r).ID

	input.Name = app.readString(qs, "name", "")
	input.Includes.Values = app.readCSV(qs, "includes", make([]string, 0))
	input.Includes.SafeList = []string{"players"}
	for _, str := range input.Includes.Values {
		if !slices.Contains(input.Includes.SafeList, str) {
			v.AddError("includes", fmt.Sprintf(`Invalid includes value.
Possible include values for teams are: "%s"`, strings.Join(input.Includes.SafeList, `", "`)))
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
	}

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 5, v)
	input.Filters.Sort = app.readString(qs, "sort", "name")
	input.Filters.SortSafeList = []string{"pin", "name", "-pin", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	teams, metadata, err := app.models.Teams.GetAll(userID, input.Name, input.Includes.Values,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "teams": teams}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	userID := app.contextGetUser(r).ID
	pin := chi.URLParam(r, "id")

	team, err := app.models.Teams.Get(userID, pin)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name       *string        `json:"name"`
		IsActive   *bool          `json:"is_active"`
		PlayerIDs  []string       `json:"player_ids"`
		PlayerNums map[string]int `json:"player_nums"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if input.Name != nil {
		if *input.Name == team.Name {
			v.AddError("name", "cannot be old name")
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		team.Name = *input.Name
	}
	if input.IsActive != nil {
		if *input.IsActive == team.IsActive {
			v.AddError("is_active", "cannot be same as old value")
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		team.IsActive = *input.IsActive
	}
	if input.PlayerIDs != nil {
		team.PlayerIDs = input.PlayerIDs
	}
	if input.PlayerNums != nil {
		team.PlayerNums = input.PlayerNums
	}

	if data.ValidateTeam(v, team); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Teams.Update(team)
	if err != nil {
		var modelValidationErr data.ModelValidationErr
		switch {
		case errors.As(err, &modelValidationErr):
			app.failedValidationResponse(w, r, modelValidationErr.Errors)
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"team": team}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	return
}

func (app *application) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	userID := app.contextGetUser(r).ID
	pin := strings.ToLower(chi.URLParam(r, "id"))

	err := app.models.Teams.Delete(userID, pin)
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
