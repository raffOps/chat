package admin

import (
	"context"
	"github.com/raffops/chat/internal/admin"
	"github.com/raffops/chat/internal/errs"
	"github.com/raffops/chat/internal/util"
	"github.com/raffops/chat/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcAdminServer struct {
	pb.AdminServer
	AdminService admin.Service
}

func (s *GrpcAdminServer) Ping(context.Context, *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Message: "ok"}, nil
}

func (s *GrpcAdminServer) ListUsers(ctx context.Context,
	_ *pb.ListUserRequest) (*pb.ListUserResponse, error) {

	users, err := s.AdminService.ListUsers(ctx)
	responseErr := errs.HttpToGrpcError(err)
	var usersProto []*pb.User
	for _, user := range users {
		usersProto = append(usersProto, &pb.User{
			Id:        user.Id,
			Name:      user.Name,
			Role:      util.StringToRole(user.Role),
			CreatedAt: timestamppb.New(user.CreatedAt),
		})
	}
	return &pb.ListUserResponse{User: usersProto}, responseErr
}

func NewGrpcAdminServer(adminService admin.Service) *GrpcAdminServer {
	return &GrpcAdminServer{AdminService: adminService}
}
