package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/kirvasik/go-homework/rok/domain"
)

const defaultWorkers = 10

type movieUseCase struct {
	repo    MovieRepository
	workers int
}

func NewMovieUseCase(repo MovieRepository) MovieUseCase {
	return &movieUseCase{
		repo:    repo,
		workers: defaultWorkers,
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

	jobs := make(chan int)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < u.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				m, err := u.repo.GetMovie(ctx, id)
				select {
				case results <- result{id: id, movie: m, err: err}:
				case <-ctx.Done():
					return
				}
			}
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
			if errors.Is(r.err, domain.ErrNotFound) {
				res.Errors = append(res.Errors, domain.FailedID{ID: r.id, Error: r.err.Error()})
			} else {
				res.Errors = append(res.Errors, domain.FailedID{ID: r.id, Error: r.err.Error()})
			}
			continue
		}
		res.Movies = append(res.Movies, r.movie)
	}

	if err := ctx.Err(); err != nil {
		return res, err
	}

	sort.Slice(res.Movies, func(i, j int) bool { return res.Movies[i].ID < res.Movies[j].ID })
	sort.Slice(res.Errors, func(i, j int) bool { return res.Errors[i].ID < res.Errors[j].ID })

	return res, nil
}

func (u *movieUseCase) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	return u.repo.GetLastMovie(ctx)
}
