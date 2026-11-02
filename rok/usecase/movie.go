package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/kirvasik/go-homework/rok/domain"
)

const (
	defaultWorkers = 10

	defaultTimeout = 10 * time.Second

	lastMovieMaxID = 99999
)

type movieUseCase struct {
	provider MovieProvider
	workers  int
}

func NewMovieUseCase(provider MovieProvider) MovieUseCase {
	return &movieUseCase{
		provider: provider,
		workers:  defaultWorkers,
	}
}

type result struct {
	id    int
	movie domain.Movie
	err   error
}

func (u *movieUseCase) LoadMovies(ctx context.Context, from, to int) (domain.LoadResult, error) {
	if from < 1 || to < from {
		return domain.LoadResult{}, fmt.Errorf("%w: from=%d, to=%d", domain.ErrInvalidRange, from, to)
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	jobs := make(chan int)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < u.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u.worker(ctx, jobs, results)
		}()
	}

	go func() {
		defer close(jobs)
		for id := from; id <= to; id++ {
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

	res := domain.LoadResult{
		Movies: make([]domain.Movie, 0, to-from+1),
	}

	for r := range results {
		if r.err != nil {
			if errors.Is(r.err, context.Canceled) || errors.Is(r.err, context.DeadlineExceeded) {
				return res, r.err
			}
			res.Errors = append(res.Errors, domain.FailedID{ID: r.id, Error: r.err.Error()})
			continue
		}
		res.Movies = append(res.Movies, r.movie)
	}

	sort.Slice(res.Movies, func(i, j int) bool { return res.Movies[i].ID < res.Movies[j].ID })
	sort.Slice(res.Errors, func(i, j int) bool { return res.Errors[i].ID < res.Errors[j].ID })

	return res, nil
}

func (u *movieUseCase) worker(ctx context.Context, jobs <-chan int, results chan<- result) {
	for id := range jobs {
		m, err := u.provider.GetMovie(ctx, id)
		select {
		case results <- result{id: id, movie: m, err: err}:
		case <-ctx.Done():
			return
		}
	}
}

func (u *movieUseCase) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	lo, hi := 1, lastMovieMaxID
	var last domain.Movie
	found := false

	for lo <= hi {
		mid := (lo + hi) / 2
		m, err := u.provider.GetMovie(ctx, mid)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				hi = mid - 1
				continue
			}
			return domain.Movie{}, err
		}
		last = m
		found = true
		lo = mid + 1
	}

	if !found {
		return domain.Movie{}, domain.ErrNotFound
	}
	return last, nil
}
