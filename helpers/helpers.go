package helpers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const InviteCodeLength int = 8

const inviteCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

var ErrTokenNotFound = errors.New("token not found")
var ErrNoGroupUpdated = errors.New("group update failed")
var ErrGroupNotDeleted = errors.New("group delete failed")
var ErrGroupMemberNotRemoved = errors.New("group member removal failed")
var ErrInvalidInviteCode = errors.New("invalid invite code")
var ErrRsvpFailed = errors.New("rsvp failed, could not find user or event")
var ErrEventNotDeleted = errors.New("event delete failed")
var ErrEventParticipantNotDeleted = errors.New("event participant removal failed")
var ErrInvalidEventParticipant = errors.New("participant must belong to event group")

type APIError struct {
	Status  int
	Message string
}

func WriteJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func WriteErr(w http.ResponseWriter, status int, msg string) {
	WriteJson(w, status, map[string]string{"error": msg})
}

func ExtractUserId(ctx context.Context) (uuid.UUID, error) {
	value := ctx.Value(auth.UserIdContextKey)

	userId, ok := value.(uuid.UUID)
	if !ok {
		return userId, errors.New("could not cast UUID")
	}

	return userId, nil
}

func GenerateNewInviteCode(length int) (string, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	result := make([]byte, length)

	for i, b := range bytes {
		result[i] = inviteCharset[int(b)%len(inviteCharset)]
	}

	return string(result), nil
}

func HandlePgxError(err error) APIError {
	log.Println("error:", err)

	if errors.Is(err, pgx.ErrNoRows) {
		return APIError{
			Status:  http.StatusNotFound,
			Message: "resource not found",
		}
	}

	if errors.Is(err, ErrTokenNotFound) {
		return APIError{
			Status:  http.StatusNotFound,
			Message: "token not found",
		}
	}

	if errors.Is(err, ErrNoGroupUpdated) {
		return APIError{
			Status:  http.StatusNotFound,
			Message: "group not found",
		}
	}

	if errors.Is(err, ErrGroupNotDeleted) {
		return APIError{
			Status:  http.StatusNotFound,
			Message: "group not found",
		}
	}
	if errors.Is(err, ErrGroupMemberNotRemoved) {
		return APIError{
			Status:  http.StatusNotFound,
			Message: "group member not found",
		}
	}
	if errors.Is(err, ErrInvalidInviteCode) {
		return APIError{
			Status:  http.StatusNotFound,
			Message: "group not found, invalid or outdated invite code",
		}
	}
	if errors.Is(err, ErrRsvpFailed) {
		return APIError{
			Status:  http.StatusBadRequest,
			Message: "rsvp operation failed",
		}
	}
	if errors.Is(err, ErrEventNotDeleted) {
		return APIError{
			Status:  http.StatusBadRequest,
			Message: "event delete failed",
		}
	}
	if errors.Is(err, ErrEventParticipantNotDeleted) {
		return APIError{
			Status:  http.StatusBadRequest,
			Message: "event participant removal failed",
		}
	}
	if errors.Is(err, ErrInvalidEventParticipant) {
		return APIError{
			Status:  http.StatusBadRequest,
			Message: "participant must belong to event group",
		}
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return APIError{
				Status:  http.StatusConflict,
				Message: "resource already exists",
			}
		case "23503": // fk_violation
			return APIError{
				Status:  http.StatusConflict,
				Message: "invalid reference",
			}
		case "23514": // check_violation
			return APIError{
				Status:  http.StatusUnprocessableEntity,
				Message: "invalid data",
			}
		case "23502": // not_null_violation
			return APIError{
				Status:  http.StatusUnprocessableEntity,
				Message: "missing required field",
			}
		case "22P02": // invalid_text_representation
			return APIError{
				Status:  http.StatusBadRequest,
				Message: "invalid input format",
			}
		}
	}

	return APIError{
		Status:  http.StatusInternalServerError,
		Message: "internal server error",
	}
}
