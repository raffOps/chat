// Code generated by mockery v2.42.1. DO NOT EDIT.

package jwt_manager

import (
	context "context"

	errs "github.com/raffops/chat/internal/errs"
	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	models "github.com/raffops/chat/internal/models"
)

// JwtManager is an autogenerated mock type for the JwtManager type
type JwtManager struct {
	mock.Mock
}

type JwtManager_Expecter struct {
	mock *mock.Mock
}

func (_m *JwtManager) EXPECT() *JwtManager_Expecter {
	return &JwtManager_Expecter{mock: &_m.Mock}
}

// AccessibleRoles provides a mock function with given fields:
func (_m *JwtManager) AccessibleRoles() map[string][]string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for AccessibleRoles")
	}

	var r0 map[string][]string
	if rf, ok := ret.Get(0).(func() map[string][]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]string)
		}
	}

	return r0
}

// JwtManager_AccessibleRoles_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AccessibleRoles'
type JwtManager_AccessibleRoles_Call struct {
	*mock.Call
}

// AccessibleRoles is a helper method to define mock.On call
func (_e *JwtManager_Expecter) AccessibleRoles() *JwtManager_AccessibleRoles_Call {
	return &JwtManager_AccessibleRoles_Call{Call: _e.mock.On("AccessibleRoles")}
}

func (_c *JwtManager_AccessibleRoles_Call) Run(run func()) *JwtManager_AccessibleRoles_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *JwtManager_AccessibleRoles_Call) Return(_a0 map[string][]string) *JwtManager_AccessibleRoles_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *JwtManager_AccessibleRoles_Call) RunAndReturn(run func() map[string][]string) *JwtManager_AccessibleRoles_Call {
	_c.Call.Return(run)
	return _c
}

// GenerateToken provides a mock function with given fields: user
func (_m *JwtManager) GenerateToken(user models.User) (string, *errs.Err) {
	ret := _m.Called(user)

	if len(ret) == 0 {
		panic("no return value specified for GenerateToken")
	}

	var r0 string
	var r1 *errs.Err
	if rf, ok := ret.Get(0).(func(models.User) (string, *errs.Err)); ok {
		return rf(user)
	}
	if rf, ok := ret.Get(0).(func(models.User) string); ok {
		r0 = rf(user)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(models.User) *errs.Err); ok {
		r1 = rf(user)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*errs.Err)
		}
	}

	return r0, r1
}

// JwtManager_GenerateToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GenerateToken'
type JwtManager_GenerateToken_Call struct {
	*mock.Call
}

// GenerateToken is a helper method to define mock.On call
//   - user models.User
func (_e *JwtManager_Expecter) GenerateToken(user interface{}) *JwtManager_GenerateToken_Call {
	return &JwtManager_GenerateToken_Call{Call: _e.mock.On("GenerateToken", user)}
}

func (_c *JwtManager_GenerateToken_Call) Run(run func(user models.User)) *JwtManager_GenerateToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(models.User))
	})
	return _c
}

func (_c *JwtManager_GenerateToken_Call) Return(_a0 string, _a1 *errs.Err) *JwtManager_GenerateToken_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *JwtManager_GenerateToken_Call) RunAndReturn(run func(models.User) (string, *errs.Err)) *JwtManager_GenerateToken_Call {
	_c.Call.Return(run)
	return _c
}

// GrpcUnaryInterceptor provides a mock function with given fields: ctx, req, info, handler
func (_m *JwtManager) GrpcUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ret := _m.Called(ctx, req, info, handler)

	if len(ret) == 0 {
		panic("no return value specified for GrpcUnaryInterceptor")
	}

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error)); ok {
		return rf(ctx, req, info, handler)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) interface{}); ok {
		r0 = rf(ctx, req, info, handler)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) error); ok {
		r1 = rf(ctx, req, info, handler)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// JwtManager_GrpcUnaryInterceptor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GrpcUnaryInterceptor'
type JwtManager_GrpcUnaryInterceptor_Call struct {
	*mock.Call
}

// GrpcUnaryInterceptor is a helper method to define mock.On call
//   - ctx context.Context
//   - req interface{}
//   - info *grpc.UnaryServerInfo
//   - handler grpc.UnaryHandler
func (_e *JwtManager_Expecter) GrpcUnaryInterceptor(ctx interface{}, req interface{}, info interface{}, handler interface{}) *JwtManager_GrpcUnaryInterceptor_Call {
	return &JwtManager_GrpcUnaryInterceptor_Call{Call: _e.mock.On("GrpcUnaryInterceptor", ctx, req, info, handler)}
}

func (_c *JwtManager_GrpcUnaryInterceptor_Call) Run(run func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler)) *JwtManager_GrpcUnaryInterceptor_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}), args[2].(*grpc.UnaryServerInfo), args[3].(grpc.UnaryHandler))
	})
	return _c
}

func (_c *JwtManager_GrpcUnaryInterceptor_Call) Return(_a0 interface{}, _a1 error) *JwtManager_GrpcUnaryInterceptor_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *JwtManager_GrpcUnaryInterceptor_Call) RunAndReturn(run func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error)) *JwtManager_GrpcUnaryInterceptor_Call {
	_c.Call.Return(run)
	return _c
}

// VerifyToken provides a mock function with given fields: tokenString
func (_m *JwtManager) VerifyToken(tokenString string) (*models.Claims, *errs.Err) {
	ret := _m.Called(tokenString)

	if len(ret) == 0 {
		panic("no return value specified for VerifyToken")
	}

	var r0 *models.Claims
	var r1 *errs.Err
	if rf, ok := ret.Get(0).(func(string) (*models.Claims, *errs.Err)); ok {
		return rf(tokenString)
	}
	if rf, ok := ret.Get(0).(func(string) *models.Claims); ok {
		r0 = rf(tokenString)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Claims)
		}
	}

	if rf, ok := ret.Get(1).(func(string) *errs.Err); ok {
		r1 = rf(tokenString)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*errs.Err)
		}
	}

	return r0, r1
}

// JwtManager_VerifyToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VerifyToken'
type JwtManager_VerifyToken_Call struct {
	*mock.Call
}

// VerifyToken is a helper method to define mock.On call
//   - tokenString string
func (_e *JwtManager_Expecter) VerifyToken(tokenString interface{}) *JwtManager_VerifyToken_Call {
	return &JwtManager_VerifyToken_Call{Call: _e.mock.On("VerifyToken", tokenString)}
}

func (_c *JwtManager_VerifyToken_Call) Run(run func(tokenString string)) *JwtManager_VerifyToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *JwtManager_VerifyToken_Call) Return(_a0 *models.Claims, _a1 *errs.Err) *JwtManager_VerifyToken_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *JwtManager_VerifyToken_Call) RunAndReturn(run func(string) (*models.Claims, *errs.Err)) *JwtManager_VerifyToken_Call {
	_c.Call.Return(run)
	return _c
}

// NewJwtManager creates a new instance of JwtManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewJwtManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *JwtManager {
	mock := &JwtManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
