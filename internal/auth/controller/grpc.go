package auth

import (
	"context"
	"github.com/raffops/chat/internal/auth"
	"github.com/raffops/chat/internal/errs"
	"github.com/raffops/chat/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net/http"
)

type GrpcAuthServer struct {
	pb.AuthServer
	AuthService auth.Service
}

func (s *GrpcAuthServer) SignUp(ctx context.Context, in *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	log.Printf("SignUp: %v", in.GetName())
	user, err := s.AuthService.SignUp(ctx, in.GetName(), in.GetPassword())
	responseErr := s.httpToGrpcError(err)
	return &pb.SignUpResponse{
		Id:        user.Id,
		Name:      user.Name,
		Role:      stringToRole(user.Role),
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, responseErr
}

func (s *GrpcAuthServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Login: %v", in.GetName())
	token, err := s.AuthService.Login(ctx, in.GetName(), in.GetPassword())
	responseErr := s.httpToGrpcError(err)
	return &pb.LoginResponse{Token: token}, responseErr
}

func (s *GrpcAuthServer) httpToGrpcError(err *errs.Err) error {
	var responseErr error
	if err != nil {
		switch err.Code {
		case http.StatusConflict:
			responseErr = status.Error(codes.AlreadyExists, err.Error())
		case http.StatusBadRequest:
			responseErr = status.Error(codes.InvalidArgument, err.Error())
		case http.StatusUnauthorized:
			responseErr = status.Error(codes.Unauthenticated, err.Error())
		default:
			responseErr = status.Error(codes.Internal, err.Error())
		}
	}
	return responseErr
}

func stringToRole(roleStr string) pb.Role {
	switch roleStr {
	case "ADMIN":
		return pb.Role_ADMIN
	case "USER":
		return pb.Role_USER
	default:
		return pb.Role_USER // default value if the string doesn't match any Role
	}
}
