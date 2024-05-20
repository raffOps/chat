package admin

import (
	"context"
	"github.com/raffops/chat/internal/errs"
	model "github.com/raffops/chat/internal/models"
)

type Service interface {
	GetUser(ctx context.Context, key, value string) (model.User, *errs.Err)
	ListUsers(ctx context.Context) ([]model.User, *errs.Err)
}
