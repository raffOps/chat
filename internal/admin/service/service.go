package admin

import (
	"context"
	"github.com/raffops/chat/internal/admin"
	"github.com/raffops/chat/internal/errs"
	"github.com/raffops/chat/internal/models"
	"github.com/raffops/chat/internal/user"
	"net/http"
)

type service struct {
	userRepo user.Repository
}

func (s service) GetUser(ctx context.Context, key, value string) (models.User, *errs.Err) {
	if key != "name" && key != "id" {
		return models.User{}, &errs.Err{Message: "invalid key", Code: http.StatusBadRequest}
	}
	return s.userRepo.GetUser(ctx, key, value)
}

func (s service) ListUsers(ctx context.Context) ([]models.User, *errs.Err) {
	//TODO implement me
	panic("implement me")
}

func NewService(userRepo user.Repository) admin.Service {
	return &service{userRepo: userRepo}
}
