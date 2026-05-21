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
		log.Fatalln(err)
		return
	}
	log.Println(".env loaded")

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_CONN"))
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("db connected, connection pool created")

	store := store.NewStore(dbpool)
	handler := hdl.NewHandler(store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.HealthHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handler.RegisterHandler)
			r.Post("/login", handler.LoginHandler)
			r.Post("/refresh", handler.RefreshHandler)
		})

		r.Group(func(r chi.Router) {
			r.Use(mid.RequireAuth)

			r.Post("/logout", handler.LogoutHandler)
		})
	})

	log.Println("routes registered")
	log.Println("server started http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
