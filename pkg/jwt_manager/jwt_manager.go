package jwt_manager

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/raffops/chat/internal/errs"
	"github.com/raffops/chat/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

type JwtManager interface {
	GenerateToken(user models.User) (string, *errs.Err)
	VerifyToken(tokenString string) (*models.Claims, *errs.Err)
	AccessibleRoles() map[string][]string
	GrpcUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
}

type jwtManager struct {
	secretKey   []byte
	jwtDuration time.Duration
	roles       map[string][]string
}

func (j *jwtManager) AccessibleRoles() map[string][]string {
	return j.roles
}

func (j *jwtManager) VerifyToken(tokenString string) (*models.Claims, *errs.Err) {
	tokenString, _ = strings.CutPrefix(tokenString, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errs.ErrUnauthorized
		}
		return j.secretKey, nil
	})
	if err != nil {
		return nil, errs.ErrUnauthorized
	}
	claims, ok := token.Claims.(*models.Claims)
	if !ok {
		return nil, errs.ErrUnauthorized
	}
	expired := claims.VerifyExpiresAt(time.Now().Unix(), true)
	if !expired {
		return nil, errs.ErrUnauthorized
	}
	return claims, nil
}

func (j *jwtManager) GenerateToken(user models.User) (string, *errs.Err) {
	claims := &models.Claims{
		Role: user.Role,
		StandardClaims: jwt.StandardClaims{
			Subject:   user.Id,
			ExpiresAt: time.Now().Add(j.jwtDuration).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokeString, errSign := token.SignedString(j.secretKey)
	if errSign != nil {
		log.Printf("failed to sign token: %v", errSign)
		return "", &errs.Err{Message: "failed to sign token", Code: http.StatusInternalServerError}
	}
	return tokeString, nil
}

func (j *jwtManager) GrpcUnaryInterceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	log.Println("--> unary interceptor: ", info.FullMethod)

	var allowedRoles []string
	var ok bool
	if allowedRoles, ok = j.roles[info.FullMethod]; !ok {
		return nil, status.Errorf(codes.Internal, "no roles are allowed for this method")
	}

	if slices.Contains(allowedRoles, "ALL") {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errs.ErrMissingMetadata
	}

	// The keys within metadata.MD are normalized to lowercase.
	// See: https://godoc.org/google.golang.org/grpc/metadata#New
	var authorization []string
	if authorization, ok = md["authorization"]; !ok {
		return nil, errs.ErrInvalidToken
	}
	claims, err := j.VerifyToken(authorization[0])
	if err != nil {
		return nil, errs.ErrInvalidToken
	}

	if !slices.Contains(allowedRoles, claims.Role) {
		return nil, errs.ErrGrpcUnauthorized
	}

	// Continue execution of handler after ensuring a valid token.
	return handler(ctx, req)
}

func NewJwtManager(secretKey []byte, jwtDuration time.Duration, accessibleRoles map[string][]string) JwtManager {
	return &jwtManager{
		secretKey:   secretKey,
		jwtDuration: jwtDuration,
		roles:       accessibleRoles,
	}
}
