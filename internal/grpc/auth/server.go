package auth

import (
	"context"
	"errors"
	ssov1 "github.com/CVSaby/proto-contracts/gen/go/sso"
	"github.com/CVSaby/sso-service/internal/domain/models"
	"github.com/CVSaby/sso-service/internal/services/auth"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auther interface {
	LoginUser(ctx context.Context, email string, password string) (token string, err error)
	RegisterUser(ctx context.Context, email string, password string, usrType models.UserType) (userID string, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auther
}

func Register(gRPC *grpc.Server, authService Auther) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: authService})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := s.validateLoginReq(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request body")
	}

	token, err := s.auth.LoginUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	var userType models.UserType

	if err := s.validateRegisterReq(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request body")
	}

	rpcUserType := req.GetUserType()
	switch rpcUserType {
	case ssov1.UserType_customer:
		userType = models.CUSTOMER
	case ssov1.UserType_seller:
		userType = models.SELLER
	default:
		userType = models.CUSTOMER
	}

	userID, err := s.auth.RegisterUser(ctx, req.GetEmail(), req.GetPassword(), userType)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) validateRegisterReq(req *ssov1.RegisterRequest) error {
	validate := validator.New()

	userReq := RegisterReqValidation{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	if err := validate.Struct(userReq); err != nil {
		return err
	}

	return nil
}

func (s *serverAPI) validateLoginReq(req *ssov1.LoginRequest) error {
	validate := validator.New()

	userReq := LoginReqValidation{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	if err := validate.Struct(userReq); err != nil {
		return err
	}

	return nil
}
