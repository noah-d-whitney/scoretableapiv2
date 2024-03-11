package main

import (
	"net/http"
	"strings"
)

func (app *application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := app.writeJSON(w, http.StatusOK, envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     app.config.version,
		},
		"cors_info": map[string]string{
			"trusted_origins": strings.Join(app.config.cors.trustedOrigins, " | "),
		},
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
