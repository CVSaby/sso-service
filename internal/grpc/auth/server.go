package auth

import (
	"context"
	ssov1 "github.com/CVSaby/proto-contracts/gen/go/sso"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auther interface {
	LoginUser(ctx context.Context, email string, password string) (token string, err error)
	RegisterUser(ctx context.Context, email string, password string) (userID int64, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auther
}

func Register(gRPC *grpc.Server, authService Auther) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: authService})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := s.auth.LoginUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// TODO: ...
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := s.auth.RegisterUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// TODO: ...
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}
