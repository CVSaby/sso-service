package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/CVSaby/sso-service/internal/config"
	"github.com/CVSaby/sso-service/internal/domain/models"
	"github.com/CVSaby/sso-service/internal/lib/jwt"
	"github.com/CVSaby/sso-service/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type Auth struct {
	log          *slog.Logger
	jwtCfg       config.JWTConfig
	userSaver    UserSaver
	userProvider UserProvider
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte, usrType models.UserType) (uid string, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user models.User, err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

func New(log *slog.Logger, userProvider UserProvider, userSaver UserSaver, jwtCfg config.JWTConfig) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		jwtCfg:       jwtCfg,
	}
}

func (a *Auth) RegisterUser(ctx context.Context, email string, password string, usrType models.UserType) (userID string, err error) {
	const op = "auth.RegisterUser"

	log := a.log.With(
		slog.String("operation", op),
	)
	log.Info("registering user")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", err, op)
	}

	id, err := a.userSaver.SaveUser(ctx, email, hash, usrType)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", slog.String("error", err.Error()))
			return "", fmt.Errorf("%w: %s", ErrUserExists, op)
		}

		log.Error("failed to save user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", err, op)
	}

	return id, nil
}

func (a *Auth) LoginUser(ctx context.Context, email string, password string) (token string, err error) {
	const op = "auth.LoginUser"

	log := a.log.With(
		slog.String("operation", op),
	)
	log.Info("user is logging")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))
			return "", fmt.Errorf("%w: %s", ErrInvalidCredentials, op)
		}

		a.log.Error("failed to get user", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", err, op)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", ErrInvalidCredentials, op)
	}

	a.log.Info("user logged in")

	token, err = jwt.NewToken(user, a.jwtCfg.AccessTokenLifeTime, a.jwtCfg.Secret)
	if err != nil {
		a.log.Error("failed to generate token", slog.String("error", err.Error()))
		return "", fmt.Errorf("%w: %s", err, op)
	}

	return token, nil
}
