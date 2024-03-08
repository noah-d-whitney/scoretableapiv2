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

func (app *application) InsertPlayer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		PrefNumber int    `json:"pref_number"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	player := &data.Player{
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		PrefNumber: input.PrefNumber,
	}

	vldtr := validator.New()
	if data.ValidatePlayer(vldtr, player); !vldtr.Valid() {
		app.failedValidationResponse(w, r, vldtr.Errors)
		return
	}

	err = app.models.Players.Insert(player)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/players/%d", player.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"player": player}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetPlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		app.notFoundResponse(w, r)
		return
	}

	player, err := app.models.Players.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"player": player}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Players.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "player successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetAllPlayers(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 5, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "first_name", "last_name", "pref_number", "-id",
		"-first_name", "-last_name", "-pref_number"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	players, metadata, err := app.models.Players.GetAll(input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "players": players}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		app.notFoundResponse(w, r)
		return
	}

	player, err := app.models.Players.Get(id)
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
		FirstName  *string `json:"first_name"`
		LastName   *string `json:"last_name"`
		PrefNumber *int    `json:"pref_number"`
		IsActive   *bool   `json:"is_active"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.FirstName != nil {
		player.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		player.LastName = *input.LastName
	}
	if input.PrefNumber != nil {
		player.PrefNumber = *input.PrefNumber
	}
	if input.IsActive != nil {
		player.IsActive = *input.IsActive
	}

	v := validator.New()
	if data.ValidatePlayer(v, player); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Players.Update(player)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"player": player}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
