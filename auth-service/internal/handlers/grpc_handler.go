package handlers

import (
	"context"
	"log/slog"

	"github.com/SneaX-23/GoServices/auth-service/internal/service"
	pb "github.com/SneaX-23/GoServices/auth-service/pkg/genproto"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	service *service.AuthService
}

func NewGRPCServer(authService *service.AuthService) *Server {
	return &Server{
		service: authService,
	}
}

func (s *Server) GetUserEmail(ctx context.Context, req *pb.UserEmailRequest) (*pb.UserEmailResponse, error) {
	userID := req.UserId
	email, err := s.service.GetEmailByID(ctx, userID)
	if err != nil {
		slog.Error("Error getting email", "err", err)
		return nil, err
	}
	return &pb.UserEmailResponse{
		Email: email,
	}, nil
}
