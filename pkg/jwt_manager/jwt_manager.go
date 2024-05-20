package jwt_manager

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/raffops/chat/internal/errs"
	"github.com/raffops/chat/internal/models"
	"log"
	"net/http"
	"time"
)

type JwtManager interface {
	GenerateToken(user models.User) (string, *errs.Err)
	VerifyToken(tokenString string) (*models.Claims, *errs.Err)
}

type jwtManager struct {
	secretKey   []byte
	jwtDuration time.Duration
}

func (j *jwtManager) VerifyToken(tokenString string) (*models.Claims, *errs.Err) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, &errs.Err{Message: "invalid token", Code: http.StatusUnauthorized}
		}
		return j.secretKey, nil
	})
	if err != nil {
		return nil, &errs.Err{Message: err.Error(), Code: http.StatusUnauthorized}
	}
	claims, ok := token.Claims.(*models.Claims)
	if !ok {
		return nil, &errs.Err{Message: "invalid claims", Code: http.StatusUnauthorized}
	}
	expired := claims.VerifyExpiresAt(time.Now().Unix(), true)
	if !expired {
		return nil, &errs.Err{Message: "token expired", Code: http.StatusUnauthorized}
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

func NewJwtManager(secretKey []byte, jwtDuration time.Duration) JwtManager {
	return &jwtManager{
		secretKey:   secretKey,
		jwtDuration: jwtDuration,
	}
}
