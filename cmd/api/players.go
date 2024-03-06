package main

import (
	"ScoreTableApi/internal/data"
	"ScoreTableApi/internal/validator"
	"fmt"
	"net/http"
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
