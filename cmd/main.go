package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

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

func main() {

	httpAddr := getenv("HTTP_ADDR", ":8080")
	grpcAddr := getenv("GRPC_ADDR", ":9090")
	movieBaseURL := getenv("MOVIE_BASE_URL", "https://homeworksite.site")

	provider := adapter.NewMovieClient(
		movieBaseURL,
		&http.Client{Timeout: 10 * time.Second},
	)
	uc := usecase.NewMovieUseCase(provider)

	httpHandler := handler.NewHTTP(uc)
	grpcHandler := handler.NewGRPC(uc)

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

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMovieServiceServer(grpcServer, grpcHandler)
	go func() {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC error: %v", err)
		}
	}()

	select {}
}
