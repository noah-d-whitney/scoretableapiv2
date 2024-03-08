package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	// Router config
	router.NotFound(app.notFoundResponse)

	// Middleware
	router.Use(app.recoverPanic)
	router.Use(app.rateLimit)

	// Healthcheck
	router.Get("/v1/healthcheck", app.HealthCheck)

	// User endpoints
	router.Post("/v1/user", app.RegisterUser)

	// Player endpoints
	router.Post("/v1/player", app.InsertPlayer)
	router.Get("/v1/player/{id}", app.GetPlayer)
	router.Get("/v1/player", app.GetAllPlayers)
	router.Delete("/v1/player/{id}", app.DeletePlayer)
	router.Patch("/v1/player/{id}", app.UpdatePlayer)

	return router
}
