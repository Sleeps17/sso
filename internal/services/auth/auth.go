package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/jwt"
	"sso/internal/storage"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:         log,
		usrProvider: userProvider,
		usrSaver:    userSaver,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appId int32) (string, error) {
	a.log.Info("try login user with email: " + email)

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))
			return "", fmt.Errorf("can't login user: %w", ErrInvalidCredentials)
		}

		a.log.Error("can't get user: %w", err.Error())

		return "", fmt.Errorf("can't get user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warn("invalid credentials", slog.String("error", err.Error()))
		return "", fmt.Errorf("can't login user: %w", ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appId)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			a.log.Warn("app not found", slog.String("error", err.Error()))
			return "", ErrInvalidAppID
		}
		a.log.Error("can't provide app", slog.String("error", err.Error()))
		return "", fmt.Errorf("can't provide app: %w", err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("can't generate jwt-token", slog.String("error", err.Error()))
		return "", fmt.Errorf("can't generate jwt-token: %w", err)
	}

	a.log.Info("user login", slog.String("email", email))

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {

	a.log.Info("try to register user", slog.String("email", email))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("can't generate password hash", slog.String("error", err.Error()))
		return 0, fmt.Errorf("can't generate password hash: %d", err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("user already exists", slog.String("email", email))
			return 0, ErrUserExists
		}
		a.log.Error("can't save new user", slog.String("email", email), slog.String("error", err.Error()))
		return 0, fmt.Errorf("can't save new user: %w", err)
	}

	a.log.Info("user registered", slog.String("email", email))

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	a.log.Info("try to check if the user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", slog.Int("userID", int(userID)))
			return false, ErrInvalidCredentials
		}
		a.log.Error("can't check if user is admin", slog.String("error", err.Error()))
		return false, fmt.Errorf("can't check if user is admin: %w", err)
	}

	a.log.Info("checked if user is admin", slog.Bool("Is_admin", isAdmin), slog.Int64("userID", userID))

	return isAdmin, nil
}
