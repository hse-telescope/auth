package facade

import (
	"context"

	"github.com/hse-telescope/auth/internal/repository/models"
)

type Storage interface {
	AddUser(ctx context.Context, username, email, password string) (int64, error)
	CheckUserByUsername(ctx context.Context, username string) (int64, string, error)
	CheckUserByEmail(ctx context.Context, email string) (int64, string, error)

	GetUserProjects(ctx context.Context, userID int64) ([]int64, error)
	CreateProjectPermission(ctx context.Context, perm models.ProjectPermission) error

	AssignRole(ctx context.Context, perm models.ProjectPermission) error
	GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error)
	DeletePermission(ctx context.Context, userID, projectID int64) error
	UpdateRole(ctx context.Context, userID, projectID int64, role string) error

	GetUserIDByUsername(ctx context.Context, username string) (int64, error)
	GetUserIDByEmail(ctx context.Context, email string) (int64, error)
	UserExists(ctx context.Context, userID int64) (bool, error)
	ProjectExists(ctx context.Context, projectID int64) (bool, error)

	ChangeUsername(ctx context.Context, username, email, password string) error
	ChangeEmail(ctx context.Context, username, email, password string) error
	ChangePassword(ctx context.Context, username, email, password string) error
}

type Facade struct {
	storage Storage
}

func New(storage Storage) Facade {
	return Facade{
		storage: storage,
	}
}

func (f Facade) AddUser(ctx context.Context, username, email, hashedPassword string) (int64, error) {
	return f.storage.AddUser(ctx, username, email, hashedPassword)
}

func (f Facade) CheckUserByUsername(ctx context.Context, username string) (int64, string, error) {
	return f.storage.CheckUserByUsername(ctx, username)
}

func (f Facade) CheckUserByEmail(ctx context.Context, email string) (int64, string, error) {
	return f.storage.CheckUserByEmail(ctx, email)
}

func (f Facade) UserExists(ctx context.Context, userID int64) (bool, error) {
	return f.storage.UserExists(ctx, userID)
}

func (f Facade) ProjectExists(ctx context.Context, projectID int64) (bool, error) {
	return f.storage.ProjectExists(ctx, projectID)
}

func (f Facade) GetUserProjects(ctx context.Context, userID int64) ([]int64, error) {
	return f.storage.GetUserProjects(ctx, userID)
}

func (f Facade) CreateProjectPermission(ctx context.Context, perm models.ProjectPermission) error {
	return f.storage.CreateProjectPermission(ctx, perm)
}

func (f Facade) GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error) {
	return f.storage.GetUserProjectRole(ctx, userID, projectID)
}

func (f Facade) AssignRole(ctx context.Context, perm models.ProjectPermission) error {
	return f.storage.AssignRole(ctx, perm)
}

func (f Facade) UpdateRole(ctx context.Context, userID, projectID int64, role string) error {
	return f.storage.UpdateRole(ctx, userID, projectID, role)
}

func (f Facade) DeletePermission(ctx context.Context, userID, projectID int64) error {
	return f.storage.DeletePermission(ctx, userID, projectID)
}

func (f Facade) GetUserIDByUsername(ctx context.Context, username string) (int64, error) {
	return f.storage.GetUserIDByUsername(ctx, username)
}

func (f Facade) GetUserIDByEmail(ctx context.Context, email string) (int64, error) {
	return f.storage.GetUserIDByEmail(ctx, email)
}

func (f Facade) ChangeUsername(ctx context.Context, username, email, password string) error {
	return f.storage.ChangeUsername(ctx, username, email, password)
}

func (f Facade) ChangeEmail(ctx context.Context, username, email, password string) error {
	return f.storage.ChangeEmail(ctx, username, email, password)
}

func (f Facade) ChangePassword(ctx context.Context, username, email, password string) error {
	return f.storage.ChangePassword(ctx, username, email, password)
}

