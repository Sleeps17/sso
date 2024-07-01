package storage

import (
	"context"
	"errors"
	"github.com/sso/internal/domain/models"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

type Storage interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
	User(ctx context.Context, email string) (models.User, error)
	App(ctx context.Context, appId int32) (models.App, error)
	IsAdmin(ctx context.Context, id int64) (bool, error)
}
