package main

import (
	"expvar"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	// Router
	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedRequest)

	// Middleware
	router.Use(app.metrics)
	router.Use(app.recoverPanic)
	router.Use(app.enableCORS)
	router.Use(app.rateLimit)
	router.Use(app.authenticate)

	// Healthcheck
	router.Get("/v1/healthcheck", app.HealthCheck)
	router.Method(http.MethodGet, "/v1/metrics", expvar.Handler())

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

	router.With(app.requireActivatedUser).Post("/v1/team", app.InsertTeam)
	router.With(app.requireActivatedUser).Delete("/v1/team/{id}", app.DeleteTeam)
	router.With(app.requireActivatedUser).Get("/v1/team/{id}", app.GetTeam)
	router.With(app.requireActivatedUser).Get("/v1/team", app.GetAllTeams)
	router.With(app.requireActivatedUser).Patch("/v1/team/{id}", app.UpdateTeam)

	router.With(app.requireActivatedUser).Post("/v1/game", app.InsertGame)
	router.With(app.requireActivatedUser).Get("/v1/game/{id}", app.GetGame)
	router.With(app.requireActivatedUser).Delete("/v1/game/{id}", app.DeleteGame)
	router.With(app.requireActivatedUser).Get("/v1/game", app.GetAllGames)
	router.With(app.requireAuthenticatedUser).Patch("/v1/game/{id}", app.UpdateGame)

	router.Get("/v1/game/start/{id}", app.StartGame)
	router.Get("/v1/game/view/{id}", app.WatchGame)

	return router
}
