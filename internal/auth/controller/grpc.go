package auth

import (
	"context"
	"github.com/raffops/chat/internal/auth"
	"github.com/raffops/chat/internal/errs"
	"github.com/raffops/chat/internal/util"
	"github.com/raffops/chat/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
)

type GrpcAuthServer struct {
	pb.AuthServer
	AuthService auth.Service
}

func (s *GrpcAuthServer) SignUp(ctx context.Context, in *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	log.Printf("SignUp: %v", in.GetName())
	user, err := s.AuthService.SignUp(ctx, in.GetName(), in.GetPassword())
	responseErr := errs.HttpToGrpcError(err)
	return &pb.SignUpResponse{
		Id:        user.Id,
		Name:      user.Name,
		Role:      util.StringToRole(user.Role),
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, responseErr
}

func (s *GrpcAuthServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Login: %v", in.GetName())
	token, err := s.AuthService.Login(ctx, in.GetName(), in.GetPassword())
	responseErr := errs.HttpToGrpcError(err)
	return &pb.LoginResponse{Token: token}, responseErr
}

func NewGrpcAuthServer(authService auth.Service) *GrpcAuthServer {
	return &GrpcAuthServer{AuthService: authService}
}
