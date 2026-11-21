package adapter

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirvasik/go-homework/rok/domain"
)

type MovieRepository struct {
	pool *pgxpool.Pool
}

func NewMovieRepository(pool *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{pool: pool}
}

func (r *MovieRepository) GetMovie(ctx context.Context, id int) (domain.Movie, error) {
	var m domain.Movie
	err := r.pool.QueryRow(ctx, `
        SELECT id, title, year, director FROM movies WHERE id = $1
    `, id).Scan(&m.ID, &m.Title, &m.Year, &m.Director)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Movie{}, domain.ErrNotFound
		}
		return domain.Movie{}, err
	}
	return m, nil
}

func (r *MovieRepository) SaveMovie(ctx context.Context, m domain.Movie) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO movies (id, title, year, director)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO NOTHING
    `, m.ID, m.Title, m.Year, m.Director)
	return err
}

func (r *MovieRepository) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	var m domain.Movie
	err := r.pool.QueryRow(ctx, `
        SELECT id, title, year, director FROM movies ORDER BY id DESC LIMIT 1
    `).Scan(&m.ID, &m.Title, &m.Year, &m.Director)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Movie{}, domain.ErrNotFound
		}
		return domain.Movie{}, err
	}
	return m, nil
}

func (r *MovieRepository) GetMaxID(ctx context.Context) (int, error) {
	var maxID int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(id), 0) FROM movies`).Scan(&maxID)
	return maxID, err
}
