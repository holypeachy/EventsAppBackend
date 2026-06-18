package helpers

import (
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type APIError struct {
	Status  int
	Message string
}

var (
	ErrTokenNotFound              = errors.New("token not found")
	ErrNoGroupUpdated             = errors.New("group update failed")
	ErrGroupNotDeleted            = errors.New("group delete failed")
	ErrGroupMemberNotRemoved      = errors.New("group member removal failed")
	ErrInvalidInviteCode          = errors.New("invalid invite code")
	ErrRsvpFailed                 = errors.New("rsvp failed, could not find user or event")
	ErrEventNotDeleted            = errors.New("event delete failed")
	ErrEventParticipantNotDeleted = errors.New("event participant removal failed")
	ErrInvalidEventParticipant    = errors.New("participant must belong to event group")
)

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
