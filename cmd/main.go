package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/kirvasik/go-homework/gen/moviesv1"
	"github.com/kirvasik/go-homework/rok/adapter"
	"github.com/kirvasik/go-homework/rok/handler"
	"github.com/kirvasik/go-homework/rok/usecase"
	"google.golang.org/grpc"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS movies (
            id BIGINT PRIMARY KEY,
            title TEXT NOT NULL,
            year INTEGER NOT NULL,
            director TEXT NOT NULL
        );
    `)
	return err
}

func main() {
	ctx := context.Background()

	httpAddr := getenv("HTTP_ADDR", ":8080")
	grpcAddr := getenv("GRPC_ADDR", ":9090")
	movieBaseURL := getenv("MOVIE_BASE_URL", "https://homeworksite.site")
	dbURL := getenv("DATABASE_URL", "postgres://user:pass@localhost:5433/movies?sslmode=disable")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := runMigrations(ctx, pool); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	client := adapter.NewMovieClient(movieBaseURL, &http.Client{Timeout: 10 * time.Second})
	repo := adapter.NewMovieRepository(pool)

	if err := usecase.SyncMovies(ctx, client, repo); err != nil {
		log.Fatalf("sync: %v", err)
	}

	uc := usecase.NewMovieUseCase(repo)

	httpHandler := handler.NewHTTP(uc)
	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           httpHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP error: %v", err)
		}
	}()

	// gRPC
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMovieServiceServer(grpcServer, handler.NewGRPC(uc))
	go func() {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC error: %v", err)
		}
	}()

	select {}
}
