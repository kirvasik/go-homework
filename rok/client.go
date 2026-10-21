package rok

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("movie not found")

type Result struct {
	ID    int
	Movie *Movie
	Err   error
}

type Config struct {
	From    int
	To      int
	Workers int
	Timeout time.Duration
}

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

type FailedID struct {
	ID    int    `json:"id"`
	Error string `json:"error"`
}

type LoadResponse struct {
	Movies []Movie    `json:"movies"`
	Errors []FailedID `json:"errors,omitempty"`
}

func movieBaseURL() string {
	if v := os.Getenv("MOVIE_BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "https://homeworksite.site"
}

func searchMovie(ctx context.Context, client *http.Client, baseURL string, id int) (*Movie, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("создание запроса: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("сетевая ошибка: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var m Movie
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("битый JSON: %w", err)
	}
	return &m, nil
}

func worker(ctx context.Context, client *http.Client, baseURL string, jobs <-chan int, results chan<- Result) {
	for id := range jobs {
		m, err := searchMovie(ctx, client, baseURL, id)
		select {
		case results <- Result{ID: id, Movie: m, Err: err}:
		case <-ctx.Done():
			return
		}
	}
}

func run(ctx context.Context, cfg Config) (LoadResponse, error) {
	client := &http.Client{Timeout: cfg.Timeout}
	baseURL := movieBaseURL()

	jobs := make(chan int)
	results := make(chan Result)

	var wg sync.WaitGroup
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, client, baseURL, jobs, results)
		}()
	}

	go func() {
		defer close(jobs)
		for id := cfg.From; id <= cfg.To; id++ {
			select {
			case jobs <- id:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	resp := LoadResponse{
		Movies: make([]Movie, 0, cfg.To-cfg.From+1),
	}

	for r := range results {
		if r.Err != nil {
			if errors.Is(r.Err, context.Canceled) || errors.Is(r.Err, context.DeadlineExceeded) {
				return resp, r.Err
			}
			resp.Errors = append(resp.Errors, FailedID{ID: r.ID, Error: r.Err.Error()})
			continue
		}
		resp.Movies = append(resp.Movies, *r.Movie)
	}

	sort.Slice(resp.Movies, func(i, j int) bool {
		return resp.Movies[i].ID < resp.Movies[j].ID
	})
	sort.Slice(resp.Errors, func(i, j int) bool {
		return resp.Errors[i].ID < resp.Errors[j].ID
	})

	return resp, nil
}

func LastMovie(ctx context.Context, client *http.Client) (*Movie, error) {
	baseURL := movieBaseURL()
	lo, hi := 1, 99999
	var last *Movie

	for lo <= hi {
		mid := (lo + hi) / 2
		m, err := searchMovie(ctx, client, baseURL, mid)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				hi = mid - 1
				continue
			}
			return nil, err
		}
		last = m
		lo = mid + 1
	}

	if last == nil {
		return nil, ErrNotFound
	}
	return last, nil
}
