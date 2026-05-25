package helpers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/auth"
)

const inviteCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

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

func GenerateNewInvite(length int) (string, error) {
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
