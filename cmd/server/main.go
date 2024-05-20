package main

import (
	"database/sql"
	authcontroller "github.com/raffops/chat/internal/auth/controller"
	authService "github.com/raffops/chat/internal/auth/service"
	"github.com/raffops/chat/internal/database"
	user "github.com/raffops/chat/internal/user/repository"
	"github.com/raffops/chat/pb"
	"github.com/raffops/chat/pkg/jwt_manager"
	"github.com/raffops/chat/pkg/password_hasher"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var (
	secretKey   = []byte("secret")
	jwtDuration = 5 * time.Duration(time.Minute)
)

func main() {
	db := database.GetPostgresConn()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Failed to close db: %v", err)
		}
	}(db)

	userRepo := user.NewPostgresRepo(db)
	passwordHasher := password_hasher.NewBcryptHasher()
	jwtManager := jwt_manager.NewJwtManager(secretKey, jwtDuration)
	authSrv := authService.NewUserService(userRepo, passwordHasher, jwtManager)

	authServer := authcontroller.GrpcAuthServer{
		AuthService: authSrv,
	}

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterAuthServer(server, &authServer)

	log.Println("Listening on port 9000...")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to server: %v", err)
	}
}
