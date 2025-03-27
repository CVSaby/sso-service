package psql

import (
	"context"
	"errors"
	"fmt"
	"github.com/CVSaby/sso-service/internal/config"
	"github.com/CVSaby/sso-service/internal/domain/models"
	"github.com/CVSaby/sso-service/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

// New creates a new connection pool to postgresql
func New(cfg config.DBConfig) (*Storage, error) {
	const op = "storage.psql.New"
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.Host, cfg.Port, cfg.DBName,
	)

	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: dbpool,
	}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte, usrType models.UserType) (uid string, err error) {
	const op = "storage.psql.SaveUser"

	query := `INSERT INTO users (email, pass_hash, user_type) VALUES ($1, $2, $3) RETURNING id`

	var user models.User

	err = s.db.QueryRow(ctx, query, email, passHash, usrType).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return "", fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			}
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return user.ID, nil
}

func (s *Storage) User(ctx context.Context, email string) (user models.User, err error) {
	const op = "storage.psql.User"

	query := `SELECT id, email, pass_hash, user_type FROM users WHERE email = $1`

	err = s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PassHash, &user.UserType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) UserByUUID(ctx context.Context, uuid uuid.UUID) (user models.User, err error) {
	const op = "storage.psql.User"

	query := `SELECT id, email, pass_hash, user_type FROM users WHERE id = $1`

	err = s.db.QueryRow(ctx, query, uuid).Scan(&user.ID, &user.Email, &user.PassHash, &user.UserType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) Close() {
	s.db.Close()
}
