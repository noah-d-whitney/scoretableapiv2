package main

import (
	"net/http"
)

func (app *application) CreatePin(w http.ResponseWriter, r *http.Request) {
	pin, err := app.models.Pins.New("team")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"pin": pin}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
