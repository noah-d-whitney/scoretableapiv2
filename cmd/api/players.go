package main

import "net/http"

func (app *application) InsertPlayer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		PrefNumber int    `json:"pref_number"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		return
	}
}
