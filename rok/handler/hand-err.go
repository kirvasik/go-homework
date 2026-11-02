package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/kirvasik/go-homework/rok/domain"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func httpStatusForError(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidRange):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusBadGateway
	}
}
