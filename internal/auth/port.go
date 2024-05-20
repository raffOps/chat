package auth

import (
	"context"
	"github.com/raffops/chat/internal/errs"
	model "github.com/raffops/chat/internal/models"
)

type Repository interface {
	GetUser(ctx context.Context, key, value string) (model.User, *errs.Err)
	CreateUser(ctx context.Context, user model.User) (model.User, *errs.Err)
}

type Service interface {
	Login(ctx context.Context, name, password string) (string, *errs.Err)
	SignUp(ctx context.Context, name, password string) (model.User, *errs.Err)
}
