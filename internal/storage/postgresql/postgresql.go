package postgresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sso/internal/domain/models"
	"github.com/sso/internal/storage"
)

const (
	UniqueViolationCode = "23505"

	defaultIDValue = 0
	userNotAdmin   = false
)

type Storage struct {
	db *pgx.Conn
}

func New(ctx context.Context, connString string) (*Storage, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("can't connect to db: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db isn't connected: %w", err)
	}

	return &Storage{
		db: conn,
	}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "postgresql.SaveUser"

	query := `INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING id`

	var userId int64
	if err := s.db.QueryRow(ctx, query, email, passHash).Scan(&userId); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolationCode {
			return defaultIDValue, storage.ErrUserExists
		}

		return defaultIDValue, fmt.Errorf("%s: can't add user: %w", op, err)
	}

	return userId, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "postgresql.User"

	query := `SELECT id, email, pass_hash FROM users WHERE email = $1`

	var user models.User

	row := s.db.QueryRow(ctx, query, email)

	if err := row.Scan(&user.ID, &user.Email, &user.PassHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("%s: can't select user: %w", op, err)
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, appId int32) (models.App, error) {
	const op = "postgresql.App"

	query := `SELECT id, name, secret FROM apps WHERE id = $1`

	var app models.App
	if err := s.db.QueryRow(ctx, query, appId).Scan(&app.ID, &app.Name, &app.Secret); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}

		return models.App{}, fmt.Errorf("%s: can't select app: %w", op, err)
	}

	return app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userId int64) (bool, error) {
	const op = "postgresql.IsAdmin"

	query := `SELECT EXISTS(SELECT 1 FROM admins WHERE user_id = $1)`

	var isAdmin bool
	if err := s.db.QueryRow(ctx, query, userId).Scan(&isAdmin); err != nil {
		return false, fmt.Errorf("%s: can't select admin: %w", op, err)
	}

	return isAdmin, nil
}
