package facade

import (
	"context"

	"github.com/hse-telesope/auth/internal/repository/models"
)

type Storage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	// GetUser(ctx context.Context, username string) (models.User, error)
	// AddUser(ctx context.Context, username string, password string) (models.User, error)
	// LoginUser(ctx context.Context, username string, password string) (models.User, error)
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
