package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	// Router
	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedRequest)

	// Middleware
	router.Use(app.recoverPanic)
	router.Use(app.rateLimit)
	router.Use(app.authenticate)

	// Healthcheck
	router.Get("/v1/healthcheck", app.HealthCheck)

	// User Endpoints
	router.Post("/v1/user", app.RegisterUser)
	router.Put("/v1/user/activate", app.ActivateUser)
	router.Post("/v1/user/login", app.LoginUser)

	// Player Endpoints
	router.Group(func(r chi.Router) {
		r.Use(app.requireActivatedUser)
		r.Post("/v1/player", app.InsertPlayer)
		r.Get("/v1/player/{id}", app.GetPlayer)
		r.Get("/v1/player", app.GetAllPlayers)
		r.Delete("/v1/player/{id}", app.DeletePlayer)
		r.Patch("/v1/player/{id}", app.UpdatePlayer)
	})

	return router
}
