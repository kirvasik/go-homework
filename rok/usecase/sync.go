package usecase

import (
	"context"
	"errors"

	"github.com/kirvasik/go-homework/rok/adapter"
	"github.com/kirvasik/go-homework/rok/domain"
)

func SyncMovies(ctx context.Context, client *adapter.MovieClient, repo *adapter.MovieRepository) error {
	lastID, err := findLastMovieID(ctx, client)
	if err != nil {
		return err
	}

	if lastID == 0 {
		return nil
	}

	startID, err := repo.GetMaxID(ctx)
	if err != nil {
		return err
	}

	limit := startID + 20
	if limit > lastID {
		limit = lastID
	}

	for id := startID + 1; id <= limit; id++ {
		m, err := client.GetMovie(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			continue
		}

		if err := repo.SaveMovie(ctx, m); err != nil {
			return err
		}
	}

	return nil
}

func findLastMovieID(ctx context.Context, client *adapter.MovieClient) (int, error) {
	lo, hi := 1, 99999
	lastID := 0
	for lo <= hi {
		mid := (lo + hi) / 2
		_, err := client.GetMovie(ctx, mid)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				hi = mid - 1
				continue
			}
			return 0, err
		}
		lastID = mid
		lo = mid + 1
	}
	return lastID, nil
}
