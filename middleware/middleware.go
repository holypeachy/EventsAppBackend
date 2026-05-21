package middleware

import "net/http"

func RequireAuth(next http.Handler) http.Handler {
	return nil
}
