package helpers

import (
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func WriteErr(w http.ResponseWriter, status int, msg string) {
	WriteJson(w, status, map[string]string{"error": msg})
}
