package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-chi/httprate"
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
	middle := mid.NewMiddleware(store, jwtSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.HealthHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Use(httprate.LimitByIP(50, time.Minute))

			r.Post("/register", handler.RegisterHandler)
			r.Post("/login", handler.LoginHandler)
			r.Post("/refresh", handler.RefreshHandler)
			r.Post("/logout", handler.LogoutHandler)
		})

		r.Group(func(r chi.Router) {
			r.Use(middle.RequireAuth)

			r.Post("/groups", handler.CreateGroupHandler)
			r.Get("/groups", handler.GetGroupsHandler)

			r.Post("/groups/join", handler.JoinGroup)

			r.Group(func(r chi.Router) {
				r.Use(middle.RequireGroupMember)

				r.Get("/groups/{groupId}", handler.GetGroupByIdHandler)
				r.Get("/groups/{groupId}/members", handler.GetGroupMembersHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(middle.RequireGroupAdmin)

				r.Post("/groups/{groupId}/invite-code/regen", handler.RegenInviteCodeHandler)
				r.Patch("/groups/{groupId}", nil)
				r.Patch("/groups/{groupId}/members/{userId}", nil)
				r.Delete("/groups/{groupId}/members/{userId}", nil)
			})

			r.Group(func(r chi.Router) {
				r.Use(middle.RequireGroupOwner)

				r.Delete("groups/{groupId}", nil)
			})
		})

	})

	log.Println("log: routes registered")
	log.Println("log: server started http://localhost:3000")
	log.Fatalln(http.ListenAndServe(":3000", r))
}
