package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"ticket-system/internal/auth"
	"ticket-system/internal/handlers"
	appmw "ticket-system/internal/middleware"
	"ticket-system/internal/store"
)

func New(s store.Store, jwtManager *auth.JWTManager) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.NoCache)

	health := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(s, jwtManager)
	ticketHandler := handlers.NewTicketHandler(s)
	authenticator := appmw.NewAuthenticator(jwtManager)

	// Public routes — no token required.
	r.Get("/health", health.Check)
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	r.Group(func(protected chi.Router) {
		protected.Use(authenticator.RequireAuth)

		protected.Post("/tickets", ticketHandler.Create)
		protected.Get("/tickets", ticketHandler.List)
		protected.Get("/tickets/{id}", ticketHandler.Get)
		protected.Patch("/tickets/{id}/status", ticketHandler.UpdateStatus)
	})

	return r
}
