package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()
	router.NotFound(app.notFoundResponse)
	router.Post("/v1/player", app.InsertPlayer)
	router.Get("/v1/player/{id}", app.GetPlayer)
	router.Get("/v1/player", app.GetAllPlayers)
	router.Delete("/v1/player/{id}", app.DeletePlayer)

	return router
}
