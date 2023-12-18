package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"sso/internal/domain/models"
	"sso/internal/storage"
)

type Storage struct {
	db *sql.DB
}

const defaultIDValue = 0
const userNotAdmin = false

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
	query := `INSERT INTO users(email, pass_hash) VALUES(?, ?)`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return defaultIDValue, fmt.Errorf("can't prepare query: %w", err)
	}

	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {

		var sqlErr sqlite3.Error

		if errors.As(err, &sqlErr) && sqlErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return defaultIDValue, storage.ErrUserExists
		}

		return defaultIDValue, fmt.Errorf("can't execute query: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return defaultIDValue, fmt.Errorf("can't get user_id: %w", err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
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

		return models.User{}, fmt.Errorf("can't select user: %w", err)
	}

	return models.User{ID: id, Email: email, PassHash: passHash}, nil
}

func (s *Storage) App(ctx context.Context, appId int32) (models.App, error) {
	query := `SELECT id, name, secret FROM apps WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, appId)

	var app models.App
	if err := row.Scan(&app.ID, &app.Name, &app.Secret); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}

		return models.App{}, fmt.Errorf("can't select app: %w", err)
	}

	return app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, id int64) (bool, error) {
	query := `SELECT is_admin FROM users WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)

	var isAdmin bool
	if err := row.Scan(&isAdmin); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return userNotAdmin, storage.ErrUserNotFound
		}

		return userNotAdmin, fmt.Errorf("can't check if user is admin: %w", err)
	}

	return isAdmin, nil
}
