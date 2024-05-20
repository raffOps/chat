package auth

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/raffops/chat/internal/auth"
	"github.com/raffops/chat/internal/errs"
	models "github.com/raffops/chat/internal/models"
	"github.com/raffops/chat/pkg/jwt_manager"
	"github.com/raffops/chat/pkg/password_hasher"
	"log"
	"net/http"
)

type service struct {
	repo       auth.Repository
	hasher     password_hasher.PasswordHasher
	jwtManager jwt_manager.JwtManager
}

func (s *service) Login(ctx context.Context, name, password string) (string, *errs.Err) {
	user, err := s.repo.GetUser(ctx, "name", name)
	if err != nil && err.Code == http.StatusInternalServerError {
		log.Printf("failed to get user: %v", err)
		return "", &errs.Err{Message: "failed to get user", Code: http.StatusInternalServerError}
	}
	if err != nil || !s.hasher.CheckPasswordHash(password, user.Password) {
		return "", &errs.Err{Message: "invalid username/password", Code: http.StatusUnauthorized}
	}

	return s.jwtManager.GenerateToken(user)
}

func (s *service) SignUp(ctx context.Context, name, password string) (models.User, *errs.Err) {
	user := models.User{Name: name, Password: password, Role: "USER"}
	validate := validator.New()
	err := validate.Struct(user)
	if err != nil {
		return models.User{}, &errs.Err{Message: err.Error(), Code: http.StatusBadRequest}
	}

	foundUser, errGet := s.repo.GetUser(ctx, "name", user.Name)
	if errGet == nil && foundUser.Name == user.Name {
		return models.User{}, &errs.Err{Message: "user already exists", Code: http.StatusConflict}
	}

	user.Password, err = s.hasher.HashPassword(user.Password)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		return models.User{}, &errs.Err{Message: "failed to hash password", Code: http.StatusInternalServerError}
	}

	createdUser, errCreate := s.repo.CreateUser(ctx, user)
	if errCreate != nil {
		return models.User{}, errCreate
	}
	return createdUser, nil
}

func NewUserService(repo auth.Repository,
	hasher password_hasher.PasswordHasher,
	jwtManager jwt_manager.JwtManager,
) auth.Service {
	return &service{repo: repo, hasher: hasher, jwtManager: jwtManager}
}
