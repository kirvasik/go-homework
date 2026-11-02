package handler

import (
	"context"
	"errors"

	pb "github.com/kirvasik/go-homework/gen/moviesv1"
	"github.com/kirvasik/go-homework/rok/domain"
	"github.com/kirvasik/go-homework/rok/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPC struct {
	pb.UnimplementedMovieServiceServer
	uc usecase.MovieUseCase
}

func NewGRPC(uc usecase.MovieUseCase) *GRPC {
	return &GRPC{uc: uc}
}

func (s *GRPC) LoadMovies(ctx context.Context, req *pb.LoadMoviesRequest) (*pb.LoadMoviesResponse, error) {
	res, err := s.uc.LoadMovies(ctx, int(req.From), int(req.To))
	if err != nil {ы
		if errors.Is(err, domain.ErrInvalidRange) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Errorf(codes.DeadlineExceeded, "timeout")
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
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
		pbErrors = append(pbErrors, &pb.FailedID{Id: int32(e.ID), Error: e.Error})
	}

	return &pb.LoadMoviesResponse{Movies: pbMovies, Errors: pbErrors}, nil
}

func (s *GRPC) GetLastMovie(ctx context.Context, _ *pb.GetLastMovieRequest) (*pb.Movie, error) {
	m, err := s.uc.GetLastMovie(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "movie not found")
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &pb.Movie{
		Id:       int32(m.ID),
		Title:    m.Title,
		Year:     int32(m.Year),
		Director: m.Director,
	}, nil
}
