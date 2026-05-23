package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/auth"
	"github.com/holypeachy/EventsAppBackend/helpers"
)

type MiddleW struct {
	jwtSecret string
}

func NewMiddleware(jwtSecret string) *MiddleW {
	return &MiddleW{jwtSecret: jwtSecret}
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
	return next
}

func (m *MiddleW) RequireGroupAdmin(next http.Handler) http.Handler {
	return next
}

func (m *MiddleW) RequireGroupOwner(next http.Handler) http.Handler {
	return next
}
