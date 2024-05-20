package auth

import (
	"context"
	"github.com/raffops/chat/internal/auth"
	"github.com/raffops/chat/internal/errs"
	mockAuth "github.com/raffops/chat/internal/mocks/auth"
	mockJwtManager "github.com/raffops/chat/internal/mocks/jwt_manager"
	mockPasswordHasher "github.com/raffops/chat/internal/mocks/password_hasher"
	"github.com/raffops/chat/internal/models"
	"github.com/raffops/chat/pkg/jwt_manager"
	"github.com/raffops/chat/pkg/password_hasher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

const (
	TestUserName     = "testuser"
	TestUserPassword = "DSIOKUFHJIOSDUYFIOSDUF2378642378623"
	TestUserRole     = "USER"
)

var (
	TestUser = models.User{
		Id:       "",
		Name:     TestUserName,
		Password: TestUserPassword,
		Role:     TestUserRole,
	}
)

func Test_service_Login(t *testing.T) {
	type fields struct {
		repo       func() auth.Repository
		hasher     func() password_hasher.PasswordHasher
		jwtManager func() jwt_manager.JwtManager
	}
	type args struct {
		ctx      context.Context
		name     string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr *errs.Err
	}{
		{
			name: "valid user and password",
			fields: fields{
				repo: func() auth.Repository {
					repo := mockAuth.NewRepository(t)
					repo.On("GetUser", context.Background(), "name", TestUserName).
						Return(TestUser, nil)
					return repo
				},
				hasher: func() password_hasher.PasswordHasher {
					hasher := mockPasswordHasher.NewPasswordHasher(t)
					hasher.On("CheckPasswordHash", TestUserPassword, TestUser.Password).
						Return(true)
					return hasher
				},
				jwtManager: func() jwt_manager.JwtManager {
					jwtManager := mockJwtManager.NewJwtManager(t)
					jwtManager.On("GenerateToken", TestUser).
						Return(mock.Anything, nil)
					return jwtManager
				},
			},
			args: args{
				ctx:      context.Background(),
				name:     TestUserName,
				password: TestUserPassword,
			},
			wantErr: nil,
		},
		{
			name: "invalid user",
			fields: fields{
				repo: func() auth.Repository {
					repo := mockAuth.NewRepository(t)
					repo.On("GetUser", context.Background(), "name", TestUserName).
						Return(models.User{}, &errs.Err{Message: "no user found", Code: http.StatusNotFound})
					return repo
				},
				hasher:     func() password_hasher.PasswordHasher { return nil },
				jwtManager: func() jwt_manager.JwtManager { return nil },
			},
			args: args{
				ctx:      context.Background(),
				name:     TestUserName,
				password: TestUserPassword,
			},
			wantErr: &errs.Err{Message: "invalid username/password", Code: http.StatusUnauthorized},
		},
		{
			name: "invalid password",
			fields: fields{
				repo: func() auth.Repository {
					repo := mockAuth.NewRepository(t)
					repo.On("GetUser", context.Background(), "name", TestUserName).
						Return(TestUser, nil)
					return repo
				},
				hasher: func() password_hasher.PasswordHasher {
					hasher := mockPasswordHasher.NewPasswordHasher(t)
					hasher.On("CheckPasswordHash", TestUserPassword, TestUser.Password).
						Return(false)
					return hasher
				},
				jwtManager: func() jwt_manager.JwtManager { return nil },
			},
			args: args{
				ctx:      context.Background(),
				name:     TestUserName,
				password: TestUserPassword,
			},
			wantErr: &errs.Err{Message: "invalid username/password", Code: http.StatusUnauthorized},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo:       tt.fields.repo(),
				hasher:     tt.fields.hasher(),
				jwtManager: tt.fields.jwtManager(),
			}
			_, gotError := s.Login(tt.args.ctx, tt.args.name, tt.args.password)
			assert.Equal(t, tt.wantErr, gotError)
		})
	}
}

func Test_service_SignUp(t *testing.T) {
	type fields struct {
		repo   func() auth.Repository
		hasher func() password_hasher.PasswordHasher
	}
	type args struct {
		ctx  context.Context
		user models.User
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantUser models.User
		wantErr  *errs.Err
	}{
		{
			name: "valid user and password",
			fields: fields{
				repo: func() auth.Repository {
					repo := mockAuth.NewRepository(t)
					repo.On("GetUser", context.Background(), "name", TestUserName).
						Return(models.User{}, nil)
					repo.On("CreateUser", context.Background(), TestUser).
						Return(TestUser, nil)
					return repo
				},
				hasher: func() password_hasher.PasswordHasher {
					hasher := mockPasswordHasher.NewPasswordHasher(t)
					hasher.On("HashPassword", TestUser.Password).Return(TestUser.Password, nil)
					return hasher
				},
			},
			args: args{
				ctx:  context.Background(),
				user: TestUser,
			},
			wantUser: TestUser,
			wantErr:  nil,
		},
		{
			name: "empty name",
			fields: fields{
				repo:   func() auth.Repository { return nil },
				hasher: func() password_hasher.PasswordHasher { return nil },
			},
			args: args{
				ctx: context.Background(),
				user: models.User{
					Name:     "",
					Password: TestUserPassword,
					Role:     TestUserRole,
				},
			},
			wantUser: models.User{},
			wantErr: &errs.Err{Message: "Key: 'User.Name' Error:Field validation for 'Name' failed on the 'required' tag",
				Code: http.StatusBadRequest},
		},
		{
			name: "empty password",
			fields: fields{
				repo:   func() auth.Repository { return nil },
				hasher: func() password_hasher.PasswordHasher { return nil },
			},
			args: args{
				ctx: context.Background(),
				user: models.User{
					Name:     TestUserName,
					Password: "",
					Role:     TestUserRole,
				},
			},
			wantUser: models.User{},
			wantErr: &errs.Err{Message: "Key: 'User.Password' Error:Field validation for 'Password' failed on the 'required' tag",
				Code: http.StatusBadRequest},
		},
		{
			name: "user already exists",
			fields: fields{
				repo: func() auth.Repository {
					repo := mockAuth.NewRepository(t)
					repo.On("GetUser", context.Background(), "name", TestUserName).
						Return(TestUser, nil)
					return repo
				},
				hasher: func() password_hasher.PasswordHasher { return nil },
			},
			args: args{
				ctx:  context.Background(),
				user: TestUser,
			},
			wantUser: models.User{},
			wantErr:  &errs.Err{Message: "user already exists", Code: http.StatusConflict},
		},
		{
			name: "hash password error",
			fields: fields{
				repo: func() auth.Repository {
					repo := mockAuth.NewRepository(t)
					repo.On("GetUser", context.Background(), "name", TestUserName).
						Return(models.User{}, nil)
					return repo
				},
				hasher: func() password_hasher.PasswordHasher {
					hasher := mockPasswordHasher.NewPasswordHasher(t)
					hasher.On("HashPassword", TestUser.Password).
						Return("", assert.AnError)
					return hasher
				},
			},
			args: args{
				ctx:  context.Background(),
				user: TestUser,
			},
			wantUser: models.User{},
			wantErr:  &errs.Err{Message: "failed to hash password", Code: http.StatusInternalServerError},
		},
		{
			name: "db create user error",
			fields: fields{
				repo: func() auth.Repository {
					repo := mockAuth.NewRepository(t)
					repo.On("GetUser", context.Background(), "name", TestUserName).
						Return(models.User{}, nil)
					repo.On("CreateUser", context.Background(), TestUser).
						Return(models.User{}, &errs.Err{Message: "internal server error", Code: http.StatusInternalServerError})
					return repo
				},
				hasher: func() password_hasher.PasswordHasher {
					hasher := mockPasswordHasher.NewPasswordHasher(t)
					hasher.On("HashPassword", TestUser.Password).Return(TestUser.Password, nil)
					return hasher
				},
			},
			args: args{
				ctx:  context.Background(),
				user: TestUser,
			},
			wantUser: models.User{},
			wantErr:  &errs.Err{Message: "internal server error", Code: http.StatusInternalServerError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo:   tt.fields.repo(),
				hasher: tt.fields.hasher(),
			}
			user, err := s.SignUp(tt.args.ctx, tt.args.user.Name, tt.args.user.Password)
			assert.Equal(t, tt.wantUser, user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
