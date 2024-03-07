package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()
	router.Post("/v1/player", app.InsertPlayer)

	return router
}
