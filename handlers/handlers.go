package handlers

import (
	"net/http"

	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/holypeachy/EventsAppBackend/store"
)

type Handler struct {
	store     *store.Store
	jwtSecret string
}

func NewHandler(store *store.Store, jwtSecret string) *Handler {
	if jwtSecret == "" {
		panic("JWT Secret is empty")
	}
	return &Handler{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "ok"})
}
