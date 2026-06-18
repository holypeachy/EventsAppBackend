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

func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) GetEventById(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, helpers.ParamEventId)

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

func (h *Handler) GetEventParticipants(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, helpers.ParamEventId)

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	parts, err := h.store.GetEventParticipants(r.Context(), eventId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, parts)
}

func (h *Handler) Rsvp(w http.ResponseWriter, r *http.Request) {
	authUserId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	eventIdString := chi.URLParam(r, helpers.ParamEventId)

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	userIdString := chi.URLParam(r, helpers.ParamUserId)

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

func (h *Handler) PatchEvent(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, helpers.ParamEventId)

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

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, helpers.ParamEventId)

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	err = h.store.DeleteEvent(r.Context(), eventId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "event deleted"})
}

func (h *Handler) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, helpers.ParamEventId)

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	userIdString := chi.URLParam(r, helpers.ParamUserId)

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		log.Println("error: invalid user id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid user id")
		return
	}

	role, err := h.store.GetEventUserRole(r.Context(), userId, eventId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}
	if role == models.EventManager || role == models.EventOwner {
		helpers.WriteErr(w, http.StatusUnauthorized, "cannot remove event manager")
		return
	}

	err = h.store.RemoveParticipant(r.Context(), eventId, userId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "user removed from event"})
}

func (h *Handler) AddParticipant(w http.ResponseWriter, r *http.Request) {
	eventIdString := chi.URLParam(r, helpers.ParamEventId)

	eventId, err := uuid.Parse(eventIdString)
	if err != nil {
		log.Println("error: invalid event id", eventId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var model models.AddParticipantsModel

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

	err = h.store.AddParticipants(r.Context(), eventId, model.ParticipantIds)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, map[string]string{"status": "event participants added"})
}

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

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

func (h *Handler) GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	userId, err := helpers.ExtractUserId(r.Context())
	if err != nil {
		log.Println("error: failed to extract user Id from ctx")
		helpers.WriteErr(w, http.StatusInternalServerError, "internal server error")
		return
	}

	groupIdString := chi.URLParam(r, helpers.ParamGroupId)

	groupId, err := uuid.Parse(groupIdString)
	if err != nil {
		log.Println("error: invalid group id", groupId)
		helpers.WriteErr(w, http.StatusBadRequest, "invalid group id")
		return
	}

	events, err := h.store.GetGroupEvents(r.Context(), userId, groupId)
	if err != nil {
		apiErr := helpers.HandlePgxError(err)
		helpers.WriteErr(w, apiErr.Status, apiErr.Message)
		return
	}

	helpers.WriteJson(w, http.StatusOK, events)
}
