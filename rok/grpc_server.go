package rok

import (
	"context"
	"net/http"
	"time"

	pb "github.com/kirvasik/go-homework/gen/moviesv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	pb.UnimplementedMovieServiceServer
}

func (s *GrpcServer) LoadMovies(ctx context.Context, req *pb.LoadMoviesRequest) (*pb.LoadMoviesResponse, error) {
	if req.From < 1 || req.To < req.From {
		return nil, status.Errorf(codes.InvalidArgument, "invalid range: from=%d, to=%d", req.From, req.To)
	}
	res, err := run(ctx, Config{
		From:    int(req.From),
		To:      int(req.To),
		Workers: 10,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	pbMovies := make([]*pb.Movie, 0, len(res.Movies))
	for _, m := range res.Movies {
		pbMovies = append(pbMovies, &pb.Movie{
			Id:       int32(m.ID),
			Title:    m.Title,
			Year:     int32(m.Year),
			Director: m.Director,
		})
	}

	pbErrors := make([]*pb.FailedID, 0, len(res.Errors))
	for _, e := range res.Errors {
		pbErrors = append(pbErrors, &pb.FailedID{
			Id:    int32(e.ID),
			Error: e.Error,
		})
	}

	return &pb.LoadMoviesResponse{
		Movies: pbMovies,
		Errors: pbErrors,
	}, nil
}

func (s *GrpcServer) GetLastMovie(ctx context.Context, req *pb.GetLastMovieRequest) (*pb.Movie, error) {

	client := &http.Client{Timeout: 10 * time.Second}
	m, err := LastMovie(ctx, client)
	if err != nil {
		return nil, err
	}

	// Конвертируем в Protobuf
	return &pb.Movie{
		Id:       int32(m.ID),
		Title:    m.Title,
		Year:     int32(m.Year),
		Director: m.Director,
	}, nil
}
