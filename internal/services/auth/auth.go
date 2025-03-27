package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CVSaby/sso-service/internal/config"
	"github.com/CVSaby/sso-service/internal/domain/models"
	"github.com/CVSaby/sso-service/internal/lib/jwt"
	"github.com/CVSaby/sso-service/internal/storage"
	"github.com/google/uuid"
	otlp "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

var tracer = otlp.Tracer("github.com/CVSaby/sso-service")

type Auth struct {
	log    *slog.Logger
	jwtCfg config.JWTConfig

	userSaver    UserSaver
	userProvider UserProvider

	producer MsgProducer
}

type MsgProducer interface {
	ProduceMessage(msg []byte) error
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte, usrType models.UserType) (uid string, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user models.User, err error)
	UserByUUID(ctx context.Context, uuid uuid.UUID) (user models.User, err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

func New(log *slog.Logger, userProvider UserProvider, userSaver UserSaver, msgProducer MsgProducer, jwtCfg config.JWTConfig) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		producer:     msgProducer,
		jwtCfg:       jwtCfg,
	}
}

func (a *Auth) RegisterUser(
	ctx context.Context,
	email string,
	FirstName string,
	LastName string,
	PhoneNumber string,
	AvatarURL string,
	password string,
	usrType models.UserType,
) (userID string, err error) {
	const op = "auth.RegisterUser"

	traceId, err := getTraceIdFromContext(ctx)
	if err != nil {
		traceId = "trace id not found in context"
	}

	ctx, span := tracer.Start(ctx, "sso-auth-register")
	defer span.End()

	log := a.log.With(
		slog.String("operation", op),
		slog.String("trace-id", traceId),
	)

	log.Info("registering user")

	// create password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		span.SetStatus(codes.Error, "unknown fail")
		span.RecordError(err)
		log.Error("failed to generate password hash", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", err, op)
	}

	// save user to db
	id, err := a.userSaver.SaveUser(ctx, email, hash, usrType)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", slog.String("error", err.Error()))
			return "", fmt.Errorf("%w: %s", ErrUserExists, op)
		}

		span.SetStatus(codes.Error, "unknown fail")
		span.RecordError(err)
		log.Error("failed to save user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", err, op)
	}

	// generate kafka message
	registerEvent := models.UserRegisteredEvent{
		UserID:      id,
		FirstName:   FirstName,
		LastName:    LastName,
		Email:       email,
		PhoneNumber: PhoneNumber,
		AvatarUrl:   AvatarURL,
		TraceID:     traceId,
	}
	kfkMessage, err := json.Marshal(registerEvent)
	if err != nil {
		log.Error("failed to marshal registerEvent", slog.String("error", err.Error()))
		span.SetStatus(codes.Error, "failed to marshal registerEvent")
		span.RecordError(err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	// produce to kafka
	err = a.producer.ProduceMessage(kfkMessage)
	if err != nil {
		log.Error("failed to produce message", slog.String("error", err.Error()))
		span.SetStatus(codes.Error, "failed to produce message")
		span.RecordError(err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (a *Auth) LoginUser(ctx context.Context, email string, password string) (token string, err error) {
	const op = "auth.LoginUser"

	traceId, err := getTraceIdFromContext(ctx)
	if err != nil {
		traceId = "trace id not found in context"
	}

	ctx, span := tracer.Start(ctx, "sso-auth-login")
	defer span.End()

	log := a.log.With(
		slog.String("operation", op),
		slog.String("trace-id", traceId),
	)

	log.Info("user is logging")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.As(err, &storage.ErrUserNotFound) {
			log.Warn("user not found", slog.String("error", err.Error()))
			return "", fmt.Errorf("%w: %s", ErrInvalidCredentials, op)
		}

		span.SetStatus(codes.Error, "unknown fail")
		span.RecordError(err)

		log.Error("failed to get user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", err, op)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", ErrInvalidCredentials, op)
	}

	log.Info("user logged in")

	token, err = jwt.NewToken(user, a.jwtCfg.AccessTokenLifeTime, a.jwtCfg.Secret)
	if err != nil {
		span.SetStatus(codes.Error, "unknown fail")
		span.RecordError(err)
		log.Error("failed to generate token", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", err, op)
	}

	return token, nil
}

func (a *Auth) IsValidJWT(ctx context.Context, token string) bool {
	const op = "auth.IsValidJWT"

	traceId, err := getTraceIdFromContext(ctx)
	if err != nil {
		traceId = "trace id not found in context"
	}

	ctx, span := tracer.Start(ctx, "sso-auth-isValidJWT")
	defer span.End()

	log := a.log.With(
		slog.String("operation", op),
		slog.String("trace-id", traceId),
	)

	log.Info("validating token")

	return jwt.IsValidJWT(token, a.jwtCfg.Secret)
}

func (a *Auth) IsUserExists(ctx context.Context, userUUID uuid.UUID) (bool, error) {
	const op = "auth.IsUserExists"

	traceId, err := getTraceIdFromContext(ctx)
	if err != nil {
		traceId = "trace id not found in context"
	}

	ctx, span := tracer.Start(ctx, "sso-auth-isUserExists")
	defer span.End()

	log := a.log.With(
		slog.String("operation", op),
		slog.String("trace-id", traceId),
	)

	log.Info("user is logging")

	_, err = a.userProvider.UserByUUID(ctx, userUUID)
	if err != nil {
		if errors.As(err, &storage.ErrUserNotFound) {
			return false, nil
		}

		span.SetStatus(codes.Error, "unknown fail")
		span.RecordError(err)

		log.Error("failed to get user", slog.String("error", err.Error()))
		return false, fmt.Errorf("%w: %s", err, op)
	}

	return true, nil
}

func getTraceIdFromContext(ctx context.Context) (string, error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return "", fmt.Errorf("%s:%s", "getTraceIdFromContext", "no trace id found")
	}

	return span.SpanContext().TraceID().String(), nil
}
