package main

import "net/http"

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int,
	message any) {
	response := envelope{"error": message}

	err := app.writeJSON(w, status, response, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
