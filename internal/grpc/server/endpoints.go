package server

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Doctor46-create/urlshort/internal/grpc/pb"
	"github.com/Doctor46-create/urlshort/internal/service"
	"github.com/Doctor46-create/urlshort/internal/auth"
)

type ShortenerServerAdapter struct {
	pb.UnimplementedShortenerServiceServer
	srvc      service.Shortener
	secretKey string
}

func NewShortenerServerAdapter(srvc service.Shortener, secretKey string) *ShortenerServerAdapter {
	return &ShortenerServerAdapter{
		srvc:      srvc,
		secretKey: secretKey,
	}
}

func (a *ShortenerServerAdapter) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, err := auth.GetUserIDFromMetadata(ctx, a.secretKey)
	if err != nil {
		return nil, err
	}

	requestID := getRequestID(ctx)

	shortKey, err := a.srvc.Shorten(req.Url, requestID, userID)
	if err != nil {
		return nil, err
	}

	return &pb.URLShortenResponse{Result: shortKey}, nil
}

func (a *ShortenerServerAdapter) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	orig, err := a.srvc.GetOriginal(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.URLExpandResponse{Result: orig}, nil
}

func (a *ShortenerServerAdapter) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, err := auth.GetUserIDFromMetadata(ctx, a.secretKey)
	if err != nil {
		return nil, err
	}

	urls, err := a.srvc.GetUserURLs(userID)
	if err != nil {
		return nil, err
	}

	resp := &pb.UserURLsResponse{}
	for _, u := range urls {
		resp.Url = append(resp.Url, &pb.URLData{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		})
	}
	return resp, nil
}
