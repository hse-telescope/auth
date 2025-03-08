package facade

import (
	"context"

	"github.com/hse-telescope/auth/internal/repository/models"
)

type Storage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	// GetUser(ctx context.Context, username string) (models.User, error)
	AddUser(ctx context.Context, username, password string) (int64, error)
	CheckUser(ctx context.Context, username string) (int64, string, error)
	// DeleteUser(ctx context.Context, username string) error
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
