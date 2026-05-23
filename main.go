package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	hdl "github.com/holypeachy/EventsAppBackend/handlers"
	mid "github.com/holypeachy/EventsAppBackend/middleware"
	"github.com/holypeachy/EventsAppBackend/store"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("error:", err)
		return
	}
	log.Println("log: .env loaded")

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_CONN"))
	if err != nil {
		log.Println("error:", err)
		return
	}
	log.Println("log: db connected, connection pool created")

	store := store.NewStore(dbpool)
	jwtSecret := os.Getenv("JWT_SECRET")
	handler := hdl.NewHandler(store, jwtSecret)
	middle := mid.NewMiddleware(jwtSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.HealthHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handler.RegisterHandler)
			r.Post("/login", handler.LoginHandler)
			r.Post("/refresh", handler.RefreshHandler)
			r.Post("/logout", handler.LogoutHandler)
		})

		r.Group(func(r chi.Router) {
			r.Use(middle.RequireAuth)

			r.Post("/groups", nil)
			r.Get("/groups", nil)
			r.Get("/groups/{groupId}", nil)

			r.Post("/groups/join", nil)
			r.Get("/groups/{groupId}/members", nil)
		})

		r.Group(func(r chi.Router) {
			r.Use(middle.RequireGroupMember)
		})

		r.Group(func(r chi.Router) {
			r.Use(middle.RequireGroupAdmin)

			r.Post("/groups/{groupId}/invite-code/regen", nil)
		})

		r.Group(func(r chi.Router) {
			r.Use(middle.RequireGroupOwner)
		})

	})

	log.Println("log: routes registered")
	log.Println("log: server started http://localhost:3000")
	log.Fatalln(http.ListenAndServe(":3000", r))
}
