package usecase

import (
	"context"

	"github.com/kirvasik/go-homework/rok/domain"
)

type MovieRepository interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
	SaveMovie(ctx context.Context, m domain.Movie) error
	GetMaxID(ctx context.Context) (int, error) // <-- Добавлен
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int) (domain.LoadResult, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}
