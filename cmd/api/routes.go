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
	router.Use(app.enableCORS)
	router.Use(app.rateLimit)
	router.Use(app.authenticate)

	// Healthcheck
	router.Get("/v1/healthcheck", app.HealthCheck)

	// User Endpoints
	router.Post("/v1/user", app.RegisterUser)
	router.Put("/v1/user/activate", app.ActivateUser)
	router.Post("/v1/user/login", app.LoginUser)

	// Player Endpoints
	router.Route("/v1/player", func(router chi.Router) {
		router.Group(func(router chi.Router) {
			router.Use(func(next http.Handler) http.Handler {
				return app.requirePermission("players:read", next)
			})
			router.Get("/{id}", app.GetPlayer)
			router.Get("/", app.GetAllPlayers)
		})

		router.Group(func(router chi.Router) {
			router.Use(func(next http.Handler) http.Handler {
				return app.requirePermission("players:write", next)
			})
			router.Post("/", app.InsertPlayer)
			router.Delete("/{id}", app.DeletePlayer)
			router.Patch("/{id}", app.UpdatePlayer)
		})
	})

	return router
}
