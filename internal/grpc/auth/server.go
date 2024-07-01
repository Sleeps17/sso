package auth

import (
	"context"
	"errors"
	ssov1 "github.com/Sleeps17/protos/gen/go/sso"
	"github.com/go-playground/validator/v10"
	"github.com/sso/internal/services/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const emptyValue = 0

type Auth interface {
	Login(ctx context.Context, email string, password string, appId int32) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	validate := validator.New()

	email := req.GetEmail()
	if err := validate.Var(email, "required,email"); err != nil {
		return nil, status.Error(codes.InvalidArgument, "email is not valid")
	}

	password := req.GetPassword()
	if len(password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "len password required will be grate then 8")
	}

	appId := req.GetAppId()
	if appId == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "invalid app id")
	}

	token, err := s.auth.Login(ctx, email, password, appId)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		if errors.Is(err, auth.ErrInvalidAppID) {
			return nil, status.Error(codes.InvalidArgument, "invalid app id")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	validate := validator.New()

	email := req.GetEmail()
	if err := validate.Var(email, "required,email"); err != nil {
		return nil, status.Error(codes.InvalidArgument, "email is not valid")
	}

	password := req.GetPassword()
	if len(password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "len password required will be grate then 8")
	}

	userID, err := s.auth.RegisterNewUser(ctx, email, password)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	userID := req.GetUserId()
	if userID == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "userID is required")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}
