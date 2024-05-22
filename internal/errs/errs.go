package errs

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

var (
	ErrUnauthorized     = &Err{Message: "Unauthorized", Code: http.StatusUnauthorized}
	ErrGrpcUnauthorized = status.Errorf(codes.Unauthenticated, "unauthorized")
	ErrInternal         = &Err{Message: "Internal server error", Code: http.StatusInternalServerError}
	ErrUserExists       = &Err{Message: "User already exists", Code: http.StatusConflict}
	ErrInvalidUser      = &Err{Message: "Invalid username/password", Code: http.StatusUnauthorized}
	ErrNotFound         = &Err{Message: "Not found", Code: http.StatusNotFound}
	ErrMissingMetadata  = status.Errorf(codes.InvalidArgument, "missing metadata")
	ErrInvalidToken     = status.Errorf(codes.Unauthenticated, "invalid token")
)

type Err struct {
	Message string
	Code    int
}

func (e *Err) Error() string {
	return e.Message
}

func HttpToGrpcError(err *Err) error {
	var responseErr error
	if err != nil {
		switch err.Code {
		case http.StatusConflict:
			responseErr = status.Error(codes.AlreadyExists, err.Error())
		case http.StatusBadRequest:
			responseErr = status.Error(codes.InvalidArgument, err.Error())
		case http.StatusUnauthorized:
			responseErr = status.Error(codes.Unauthenticated, err.Error())
		default:
			responseErr = status.Error(codes.Internal, err.Error())
		}
	}
	return responseErr
}
