package facade

import (
	"context"
	"time"

	"github.com/hse-telescope/auth/internal/repository/models"
)

type Storage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	AddUser(ctx context.Context, username, password string) (int64, error)
	CheckUser(ctx context.Context, username string) (int64, string, error)
	GetRefreshToken(ctx context.Context, refreshToken string) (models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, refreshToken string) error
	DeleteExpiredRefreshTokens(ctx context.Context) error
	CreateRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
}

type Facade struct {
	storage Storage
}

func New(storage Storage) Facade {
	return Facade{
		storage: storage,
	}
}

func (f Facade) GetUsers(ctx context.Context) ([]models.User, error) {
	return f.storage.GetUsers(ctx)
}

func (f Facade) AddUser(ctx context.Context, username, hashedPassword string) (int64, error) {
	return f.storage.AddUser(ctx, username, hashedPassword)
}

func (f Facade) CheckUser(ctx context.Context, username string) (int64, string, error) {
	return f.storage.CheckUser(ctx, username)
}

func (f Facade) GetRefreshToken(ctx context.Context, refreshToken string) (models.RefreshToken, error) {
	return f.storage.GetRefreshToken(ctx, refreshToken)
}

func (f Facade) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	return f.storage.DeleteRefreshToken(ctx, refreshToken)
}

func (f Facade) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return f.storage.DeleteExpiredRefreshTokens(ctx)
}

func (f Facade) CreateRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	return f.storage.CreateRefreshToken(ctx, userID, refreshToken, expiresAt)
}
