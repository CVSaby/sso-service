package auth

import (
	"context"
	"errors"
	"fmt"
	ssov1 "github.com/CVSaby/proto-contracts/gen/go/sso"
	"github.com/CVSaby/sso-service/internal/domain/models"
	"github.com/CVSaby/sso-service/internal/services/auth"
	"github.com/CVSaby/sso-service/pkg/metric"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	otlp "go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var tracer = otlp.Tracer("github.com/CVSaby/sso-service")

type Auther interface {
	LoginUser(ctx context.Context, email string, password string) (token string, err error)
	RegisterUser(
		ctx context.Context,
		email string,
		FirstName string,
		LastName string,
		PhoneNumber string,
		AvatarURL string,
		password string,
		usrType models.UserType,
	) (userID string, err error)
	IsValidJWT(ctx context.Context, token string) bool
	IsUserExists(ctx context.Context, userUUID uuid.UUID) (bool, error)
}

type serverAPI struct {
	ssov1.AuthServer
	auth Auther
}

func Register(gRPC *grpc.Server, authService Auther) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: authService})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	start := time.Now()

	ctx, span := tracer.Start(ctx, "sso-grpc-login")
	defer span.End()

	appmetrics.RequestsCounter.Add(ctx, 1)

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

	appmetrics.RequestLatency.Record(ctx, time.Since(start).Seconds())

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	start := time.Now()

	ctx, span := tracer.Start(ctx, "sso-grpc-register")
	defer span.End()
	appmetrics.RequestsCounter.Add(ctx, 1)

	if err := s.validateRegisterReq(req); err != nil {
		fmt.Println(err.Error())
		return nil, status.Error(codes.InvalidArgument, "invalid request body")
	}

	var userType models.UserType
	rpcUserType := req.GetUserType()
	switch rpcUserType {
	case ssov1.Role_user:
		userType = models.USER
	case ssov1.Role_admin:
		userType = models.ADMIN
	default:
		userType = models.USER
	}

	userID, err := s.auth.RegisterUser(
		ctx,
		req.GetEmail(),
		req.GetFirstName(),
		req.GetLastName(),
		req.GetPhoneNumber(),
		req.GetAvatarUrl(),
		req.GetPassword(),
		userType,
	)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	appmetrics.RequestLatency.Record(ctx, time.Since(start).Seconds())
	return &ssov1.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) IsUserExists(ctx context.Context, req *ssov1.IsUserExistsRequest) (*ssov1.IsUserExistsResponse, error) {
	start := time.Now()

	ctx, span := tracer.Start(ctx, "sso-grpc-isUserExists")
	defer span.End()

	appmetrics.RequestsCounter.Add(ctx, 1)

	err := uuid.Validate(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userUUID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	exists, err := s.auth.IsUserExists(ctx, userUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	appmetrics.RequestLatency.Record(ctx, time.Since(start).Seconds())
	return &ssov1.IsUserExistsResponse{Exists: exists}, nil
}

func (s *serverAPI) IsValidJWT(ctx context.Context, req *ssov1.IsValidJWTRequest) (*ssov1.IsValidJWTResponse, error) {
	start := time.Now()

	ctx, span := tracer.Start(ctx, "sso-grpc-isValidJWT")
	defer span.End()

	appmetrics.RequestsCounter.Add(ctx, 1)

	valid := s.auth.IsValidJWT(ctx, req.GetJwt())

	appmetrics.RequestLatency.Record(ctx, time.Since(start).Seconds())

	return &ssov1.IsValidJWTResponse{
		Exists: valid,
	}, nil
}

func (s *serverAPI) validateRegisterReq(req *ssov1.RegisterRequest) error {
	validate := validator.New()

	userReq := RegisterReqValidation{
		Email:       req.GetEmail(),
		FirstName:   req.GetFirstName(),
		LastName:    req.GetLastName(),
		PhoneNumber: req.GetPhoneNumber(),
		AvatarUrl:   req.GetAvatarUrl(),
		Password:    req.GetPassword(),
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
