package main

import "net/http"

func (app *application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusOK, envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     app.config.version,
		},
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
