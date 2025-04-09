package facade

import (
	"context"

	"github.com/hse-telescope/auth/internal/repository/models"
)

type Storage interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	AddUser(ctx context.Context, username, password string) (int64, error)
	CheckUser(ctx context.Context, username string) (int64, string, error)
	CreateProjectPermission(ctx context.Context, perm models.ProjectPermission) error
	GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error)
	UserExists(ctx context.Context, userID int64) (bool, error)
	ProjectExists(ctx context.Context, projectID int64) (bool, error)
	DeletePermission(ctx context.Context, userID, projectID int64) error
	UpdateRole(ctx context.Context, userID, projectID int64, role string) error
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

func (f Facade) CreateProjectPermission(ctx context.Context, perm models.ProjectPermission) error {
	return f.storage.CreateProjectPermission(ctx, perm)
}

func (f Facade) GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error) {
	return f.storage.GetUserProjectRole(ctx, userID, projectID)
}

func (f Facade) UserExists(ctx context.Context, userID int64) (bool, error) {
	return f.storage.UserExists(ctx, userID)
}

func (f Facade) ProjectExists(ctx context.Context, projectID int64) (bool, error) {
	return f.storage.ProjectExists(ctx, projectID)
}

func (f Facade) DeletePermission(ctx context.Context, userID, projectID int64) error {
	return f.storage.DeletePermission(ctx, userID, projectID)
}

func (f Facade) UpdateRole(ctx context.Context, userID, projectID int64, role string) error {
	return f.storage.UpdateRole(ctx, userID, projectID, role)
}
