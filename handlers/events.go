package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/holypeachy/EventsAppBackend/models"
)

func (h *Handler) GetEventsHandler(w http.ResponseWriter, r *http.Request) {

}
func (h *Handler) GetEventByIdHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) GetEventParticipantsHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) RsvpHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) PatchEventHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) RemoveParticipantHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) AddParticipantHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	var model models.EventModel

	err = json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}

	err = model.Validate()
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(string(model.Status)) == "" {
		model.Status = models.EventRsvpOpen
	}
	eventRow, err := h.store.CreateEvent(r.Context(), groupId, userId, model)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusCreated, eventRow)
}

func (h *Handler) GetGroupEventsHandler(w http.ResponseWriter, r *http.Request) {
	// only if user is in event
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}
	log.Println(userId)

	groupIdString := chi.URLParam(r, "groupId")

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	events, err := h.store.GetGroupEvents(r.Context(), groupId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, events)
}
