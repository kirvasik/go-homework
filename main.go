package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

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

const baseURL = "https://homeworksite.site"

func Parseflag() (Config, error) {
	var (
		from    = flag.Int("from", -1, "id первого фильма")
		to      = flag.Int("to", -1, "id последнего фильма")
		workers = flag.Int("workers", 10, "ну сколько рабочих")
		timeout = flag.Duration("timeout", 5*time.Second, "ну таймаут запроса")
	)
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	var toSet, fromSet bool
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "from":
			fromSet = true
		case "to":
			toSet = true
		}
	})
	if !fromSet {
		return Config{}, errors.New("флаг --from обязателен")
	}
	if !toSet {
		return Config{}, errors.New("флаг --to обязателен")
	}
	if *from < 0 {
		return Config{}, fmt.Errorf("--from должен быть >= 0, получено %d", *from)
	}
	if *to < 0 {
		return Config{}, fmt.Errorf("--to должен быть >= 0, получено %d", *to)
	}
	if *from > *to {
		return Config{}, fmt.Errorf("--from (%d) не может быть больше --to (%d)", *from, *to)
	}
	if *workers <= 0 {
		return Config{}, fmt.Errorf("--workers должен быть > 0, получено %d", *workers)
	}
	if *timeout <= 0 {
		return Config{}, fmt.Errorf("--timeout должен быть > 0, получено %v", *timeout)
	}

	return Config{
		From:    *from,
		To:      *to,
		Workers: *workers,
		Timeout: *timeout,
	}, nil

}
func serchMovie(ctx context.Context, client *http.Client, id int) (*Movie, error) {
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var m Movie
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("битый JSON: %w", err)
	}
	return &m, nil
}
func worker(ctx context.Context, client *http.Client, jobs <-chan int, results chan<- Result) {
	for id := range jobs {
		m, err := serchMovie(ctx, client, id)
		select {
		case results <- Result{ID: id, Movie: m, Err: err}:
		case <-ctx.Done():
			return
		}
	}
}
func run(ctx context.Context, cfg Config) {
	client := &http.Client{Timeout: cfg.Timeout}

	jobs := make(chan int)
	results := make(chan Result)

	var wg sync.WaitGroup

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, client, jobs, results)
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

	for r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "id=%d: ошибка: %v\n", r.ID, r.Err)
			continue
		}
		fmt.Printf("%d — %s — %d — %s\n",
			r.Movie.ID, r.Movie.Title, r.Movie.Year, r.Movie.Director)
	}
}
func main() {
	cfg, err := Parseflag()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка: %v\n", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	run(ctx, cfg)
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "прервано пользователем")
		stop()
		os.Exit(130)
	}
}
