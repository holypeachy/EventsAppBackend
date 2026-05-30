package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/auth"
	"github.com/holypeachy/EventsAppBackend/helpers"
	"github.com/holypeachy/EventsAppBackend/models"
	"github.com/holypeachy/EventsAppBackend/store"
)

type MiddleW struct {
	store     *store.Store
	jwtSecret string
}

func NewMiddleware(store *store.Store, jwtSecret string) *MiddleW {
	return &MiddleW{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (m *MiddleW) RequireAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			helpers.WriteErr(w, http.StatusUnauthorized, "missing auth header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			helpers.WriteErr(w, http.StatusUnauthorized, "invalid auth header")
			return
		}

		tokenString := parts[1]

		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(m.jwtSecret), nil
		},
		)

		if err != nil || !token.Valid {
			helpers.WriteErr(w, http.StatusUnauthorized, "invalid token")
			return
		}

		userId, err := uuid.Parse(claims.Subject)
		if err != nil {
			log.Println("error:", err)
			helpers.WriteErr(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(
			r.Context(),
			auth.UserIdContextKey,
			userId,
		)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m *MiddleW) RequireGroupMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		doesBelong, err := m.store.DoesUserBelongToGroup(r.Context(), userId, groupId)
		if err != nil {
			apiErr := helpers.HandlePgxError(err)
			helpers.WriteErr(w, apiErr.Status, apiErr.Message)
			return
		}

		if !doesBelong {
			helpers.WriteErr(w, http.StatusUnauthorized, "user does not belong to group")
		} else {
			next.ServeHTTP(w, r)
			log.Println("log: middleware, user is member of group")
		}
	})
}

func (m *MiddleW) RequireGroupAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		role, err := m.store.GetUserRoleInGroup(r.Context(), userId, groupId)
		if err != nil {
			apiErr := helpers.HandlePgxError(err)
			helpers.WriteErr(w, apiErr.Status, apiErr.Message)
			return
		}

		if role == models.Admin || role == models.Owner {
			next.ServeHTTP(w, r)
			log.Println("log: middleware, user is admin of group")
		} else {
			helpers.WriteErr(w, http.StatusUnauthorized, "user is not group admin")
		}
	})
}

func (m *MiddleW) RequireGroupOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		role, err := m.store.GetUserRoleInGroup(r.Context(), userId, groupId)
		if err != nil {
			apiErr := helpers.HandlePgxError(err)
			helpers.WriteErr(w, apiErr.Status, apiErr.Message)
			return
		}

		if role != models.Owner {
			helpers.WriteErr(w, http.StatusUnauthorized, "user is not group owner")
		} else {
			next.ServeHTTP(w, r)
			log.Println("log: middleware, user is owner of group")
		}
	})
}

func (m *MiddleW) RequireEventParticipant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, err := helpers.ExtractUserId(r.Context())
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

		exists, err := m.store.IsUserPartOfEvent(r.Context(), userId, eventId)
		if err != nil {
			apiErr := helpers.HandlePgxError(err)
			helpers.WriteErr(w, apiErr.Status, apiErr.Message)
			return
		}

		if !exists {
			helpers.WriteErr(w, http.StatusUnauthorized, "user is not event participant")
			return
		}
		next.ServeHTTP(w, r)
		log.Println("log: middleware, user is event participant")
	})
}

func (m *MiddleW) RequireEventAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, err := helpers.ExtractUserId(r.Context())
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

		role, err := m.store.GetEventUserRole(r.Context(), userId, eventId)
		if err != nil {
			apiErr := helpers.HandlePgxError(err)
			helpers.WriteErr(w, apiErr.Status, apiErr.Message)
			return
		}

		if role != models.EventAdmin {
			helpers.WriteErr(w, http.StatusUnauthorized, "user is not event admin")
			return
		}
		next.ServeHTTP(w, r)
		log.Println("log: middleware, user is event admin")
	})
}
