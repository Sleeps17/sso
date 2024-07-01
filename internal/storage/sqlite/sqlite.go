package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/sso/internal/domain/models"
	"github.com/sso/internal/storage"
)

type Storage struct {
	db *sql.DB
}

const (
	defaultIDValue = 0
	userNotAdmin   = false
)

func New(storagePath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("can't open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db isn't connected: %w", err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "sqlite.SaveUser"

	query := `INSERT INTO users(email, pass_hash) VALUES(?, ?)`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return defaultIDValue, fmt.Errorf("%s: can't prepare query: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {

		var sqlErr sqlite3.Error

		if errors.As(err, &sqlErr) && errors.Is(sqlErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return defaultIDValue, storage.ErrUserExists
		}

		return defaultIDValue, fmt.Errorf("%s: can't execute query: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return defaultIDValue, fmt.Errorf("%s: can't get user_id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "sqlite.User"

	query := `SELECT id, email, pass_hash FROM users WHERE email = ?`

	row := s.db.QueryRowContext(ctx, query, email)

	var (
		id       int64
		passHash []byte
	)

	if err := row.Scan(&id, &email, &passHash); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("%s: can't select user: %w", op, err)
	}

	return models.User{ID: id, Email: email, PassHash: passHash}, nil
}

func (s *Storage) App(ctx context.Context, appId int32) (models.App, error) {
	const op = "sqlite.App"

	query := `SELECT id, name, secret FROM apps WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, appId)

	var app models.App
	if err := row.Scan(&app.ID, &app.Name, &app.Secret); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}

		return models.App{}, fmt.Errorf("%s: can't select app: %w", op, err)
	}

	return app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userId int64) (bool, error) {
	const op = "sqlite.IsAdmin"

	query := `SELECT EXISTS(SELECT 1 FROM admins WHERE user_id = ?)`

	row := s.db.QueryRowContext(ctx, query, userId)

	var isAdmin bool
	if err := row.Scan(&isAdmin); err != nil {
		return userNotAdmin, fmt.Errorf("%s: can't check if user is admin: %w", op, err)
	}

	return isAdmin, nil
}
