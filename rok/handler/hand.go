package handler

import (
	"context"
	"embed"
	"errors"
	"net/http"
	"strconv"

	"github.com/kirvasik/go-homework/rok/domain"
	"github.com/kirvasik/go-homework/rok/usecase"
)

//go:embed swagger.html openapi.json
var staticFS embed.FS

type movieResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

func toMovieResponse(m domain.Movie) movieResponse {
	return movieResponse{
		ID:       m.ID,
		Title:    m.Title,
		Year:     m.Year,
		Director: m.Director,
	}
}

type HTTP struct {
	uc usecase.MovieUseCase
}

func NewHTTP(uc usecase.MovieUseCase) http.Handler {
	h := &HTTP{uc: uc}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /load-movies", h.loadMovies)
	mux.HandleFunc("GET /last-movie", h.getLastMovie)
	mux.HandleFunc("GET /swagger/", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFS.ReadFile("swagger.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFS.ReadFile("openapi.json")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
	return mux
}

func (h *HTTP) loadMovies(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromStr, toStr := q.Get("from"), q.Get("to")
	if fromStr == "" || toStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "from and to are required"})
		return
	}
	from, err := strconv.Atoi(fromStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid from"})
		return
	}
	to, err := strconv.Atoi(toStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid to"})
		return
	}

	res, err := h.uc.LoadMovies(r.Context(), from, to)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "timeout"})
		case errors.Is(err, context.Canceled):
			return
		default:
			writeJSON(w, httpStatusForError(err), map[string]string{"error": err.Error()})
		}
		return
	}

	out := make([]movieResponse, 0, len(res.Movies))
	for _, m := range res.Movies {
		out = append(out, toMovieResponse(m))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *HTTP) getLastMovie(w http.ResponseWriter, r *http.Request) {
	m, err := h.uc.GetLastMovie(r.Context())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, httpStatusForError(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, toMovieResponse(m))
}
