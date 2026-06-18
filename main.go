package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	hdl "github.com/holypeachy/EventsAppBackend/handlers"
	mid "github.com/holypeachy/EventsAppBackend/middleware"
	sto "github.com/holypeachy/EventsAppBackend/store"

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

	store := sto.NewStore(dbpool)
	jwtSecret := os.Getenv("JWT_SECRET")

	handler := hdl.NewHandler(store, jwtSecret)
	middle := mid.NewMiddleware(store, jwtSecret)

	r := chi.NewRouter()
	registerRoutes(r, handler, middle)

	log.Println("log: routes registered")
	log.Println("log: server started http://localhost:3000")
	log.Fatalln(http.ListenAndServe(":3000", r))
}
