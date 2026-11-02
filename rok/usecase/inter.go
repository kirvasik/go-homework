package usecase

import (
	"context"

	"github.com/kirvasik/go-homework/rok/domain"
)

type MovieProvider interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int) (domain.LoadResult, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}
