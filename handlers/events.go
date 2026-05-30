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
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	events, err := h.store.GetEvents(r.Context(), userId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, events)

}
func (h *Handler) GetEventByIdHandler(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, "eventId")

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.store.GetEventById(r.Context(), eventId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, event)
}

func (h *Handler) GetEventParticipantsHandler(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, "eventId")

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	participants, err := h.store.GetEventParticipants(r.Context(), eventId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, participants)
}

func (h *Handler) RsvpHandler(w http.ResponseWriter, r *http.Request) {
	authUserId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	eventIdString := chi.URLParam(r, "eventId")

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	userIdString := chi.URLParam(r, "userId")

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		log.Println("error: invalid user id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if authUserId != userId {
		helpers.WriteErr(w, http.StatusUnauthorized, "userId parameter does not match authenticated user")
		return
	}

	var model models.RsvpModel
	err = json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}

	err = model.Validate()
	if err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	part, err := h.store.Rsvp(r.Context(), userId, eventId, models.ParticipantStatus(model.Status))
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, part)
}

func (h *Handler) PatchEventHandler(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, "eventId")

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var model models.UpdateEventModel
	err = json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		log.Println("error:", err)
		helpers.WriteErr(w, http.StatusBadRequest, "malformed request")
		return
	}
	err = model.Validate()
	if err != nil {
		helpers.WriteErr(w, http.StatusBadRequest, err.Error())
		return
	}

	event, err := h.store.UpdateEventInfo(r.Context(), eventId, model)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, event)
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

	var model models.EventModelDto

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
