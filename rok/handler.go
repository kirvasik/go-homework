package rok

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

type API struct{}

func NewHandler() http.Handler {
	api := &API{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", api.list)
	mux.HandleFunc("GET /last-movie", api.get)
	mux.HandleFunc("GET /swagger/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger.html")
	})
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "openapi.json")
	})
	return mux
}

func (a *API) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	fromStr, toStr := q.Get("from"), q.Get("to")
	if fromStr == "" || toStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "from and to are required"})
		return
	}
	from, err := strconv.Atoi(fromStr)
	if err != nil || from < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid from"})
		return
	}
	to, err := strconv.Atoi(toStr)
	if err != nil || to < from {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid to"})
		return
	}

	result, err := run(r.Context(), Config{
		From: from, To: to, Workers: 10, Timeout: 10 * time.Second,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "timeout"})
		case errors.Is(err, context.Canceled):

		default:
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		}
		return
	}
	writeJSON(w, http.StatusOK, result.Movies)
}

func (a *API) get(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{Timeout: 10 * time.Second}
	m, err := LastMovie(r.Context(), client)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, m)
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}
