package server

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Doctor46-create/urlshort/internal/grpc/pb"
)

type ShortenerServer interface {
	ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error)
	ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error)
	ListUserURLs(ctx context.Context, req *emptypb.Empty) (*pb.UserURLsResponse, error)
}
