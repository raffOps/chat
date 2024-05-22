package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"fmt"
	admin_controller "github.com/raffops/chat/internal/admin/controller"
	admin_service "github.com/raffops/chat/internal/admin/service"
	auth_controller "github.com/raffops/chat/internal/auth/controller"
	auth_service "github.com/raffops/chat/internal/auth/service"
	"github.com/raffops/chat/internal/database"
	user "github.com/raffops/chat/internal/user/repository"
	"github.com/raffops/chat/pb"
	"github.com/raffops/chat/pkg/jwt_manager"
	"github.com/raffops/chat/pkg/password_hasher"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"os"
	"time"
)

const (
	serverCertFile   = "cert/server-cert.pem"
	serverKeyFile    = "cert/server-key.pem"
	clientCACertFile = "cert/ca-cert.pem"
)

var (
	secretKey       = []byte(os.Getenv("SECRET_KEY"))
	jwtDuration     = 600 * time.Duration(time.Minute)
	accessibleRoles = map[string][]string{
		"/auth.Auth/SignUp":      {"ALL"},
		"/auth.Auth/Login":       {"ALL"},
		"/admin.Admin/Ping":      {"ADMIN"},
		"/admin.Admin/ListUsers": {"ADMIN"},
	}
)

func main() {
	enableTLS := flag.Bool("tls", true, "enable SSL/TLS")

	db := database.GetPostgresConn()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Failed to close db: %v", err)
		}
	}(db)

	userRepo := user.NewPostgresRepo(db)
	passwordHasher := password_hasher.NewBcryptHasher()
	jwtManager := jwt_manager.NewJwtManager(secretKey, jwtDuration, accessibleRoles)
	authService := auth_service.NewUserService(userRepo, passwordHasher, jwtManager)
	authServer := auth_controller.NewGrpcAuthServer(authService)

	adminService := admin_service.NewService(userRepo)
	adminServer := admin_controller.NewGrpcAdminServer(adminService)
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(jwtManager.GrpcUnaryInterceptor),
	}

	if *enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Panicf("cannot load TLS credentials: %v", err)
		}

		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	}

	server := grpc.NewServer(serverOptions...)
	pb.RegisterAuthServer(server, authServer)
	pb.RegisterAdminServer(server, adminServer)

	log.Println("Listening on port 9000...")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Carregar o certificado da CA que assinou o certificado do cliente
	pemClientCA, err := os.ReadFile(clientCACertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("falha ao adicionar o certificado da CA do cliente")
	}

	// Carregar o certificado e a chave privada do servidor
	serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
	if err != nil {
		return nil, err
	}

	// Configurar as credenciais do TLS e retornar
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			log.Println("Verifying client certificate...")
			if len(verifiedChains) == 0 {
				return fmt.Errorf("no verified chains")
			}
			// Log additional details for debugging
			for _, chain := range verifiedChains {
				for _, cert := range chain {
					log.Printf("Client Cert: Subject=%s, Issuer=%s\n", cert.Subject, cert.Issuer)
				}
			}
			return nil
		},
	}

	return credentials.NewTLS(config), nil
}
