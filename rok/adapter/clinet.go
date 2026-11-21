package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kirvasik/go-homework/rok/domain"
)

type movieDTO struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type MovieClient struct {
	baseURL string
	client  *http.Client
}

func NewMovieClient(baseURL string, client *http.Client) *MovieClient {
	return &MovieClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

func (c *MovieClient) GetMovie(ctx context.Context, id int) (domain.Movie, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.Movie{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return domain.Movie{}, err
		}
		return domain.Movie{}, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Movie{}, domain.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Movie{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var dto movieDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return domain.Movie{}, fmt.Errorf("decode JSON: %w", err)
	}

	return domain.Movie{
		ID:       dto.ID,
		Title:    dto.Title,
		Year:     dto.Year,
		Director: dto.Director,
	}, nil
}
