package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/hse-telescope/auth/internal/auth"
	"github.com/hse-telescope/auth/internal/repository/models"
	"github.com/olegdayo/omniconv"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	AddUser(ctx context.Context, username, password string) (int64, error)
	CheckUser(ctx context.Context, username string) (int64, string, error)

	UserExists(ctx context.Context, userID int64) (bool, error)
	ProjectExists(ctx context.Context, projectID int64) (bool, error)
	CreateProjectPermission(ctx context.Context, perm models.ProjectPermission) error
	GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error)
	UpdateRole(ctx context.Context, userID, projectID int64, role string) error
	DeletePermission(ctx context.Context, userID, projectID int64) error
}

type Provider struct {
	repository Repository
}

func New(repository Repository) Provider {
	return Provider{repository: repository}
}

var (
	ErrPermissionExists   = errors.New("permission already exists")
	ErrOnlyOwnerCanAssign = errors.New("only owner can assign roles")
	ErrInvalidRole        = errors.New("invalid role")
	ErrUserNotFound       = errors.New("user not found")
	ErrProjectNotFound    = errors.New("project not found")
	ErrPermissionNotFound = errors.New("permission not found")
)

// GetAllUsers

func (p Provider) GetUsers(ctx context.Context) ([]User, error) {
	users, err := p.repository.GetUsers(ctx)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return omniconv.ConvertSlice(users, DBUser2ProviderUser), nil
}

// Register

func (p Provider) RegisterUser(ctx context.Context, username, password string) (TokenPair, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to hash password: %w", err)
	}

	_, _, err = p.repository.CheckUser(ctx, username)
	if err == nil {
		return TokenPair{}, fmt.Errorf("user already exists")
	} else if err != sql.ErrNoRows {
		return TokenPair{}, fmt.Errorf("failed to check user: %w", err)
	}

	userID, err := p.repository.AddUser(ctx, username, string(hashedPassword))
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to add user: %w", err)
	}

	return p.GenerateTokens(ctx, userID)
}

// Login

func (p Provider) LoginUser(ctx context.Context, username, password string) (TokenPair, error) {
	userID, hashedPassword, err := p.repository.CheckUser(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return TokenPair{}, fmt.Errorf("user not found")
		}
		fmt.Println(err.Error())
		return TokenPair{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		fmt.Println(err.Error())
		return TokenPair{}, fmt.Errorf("incorrect password")
	}

	return p.GenerateTokens(ctx, userID)
}

// Refresh tokens

func (p Provider) RefreshTokens(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := auth.ValidateToken(refreshToken)
	if err != nil {
		return TokenPair{}, fmt.Errorf("invalid refresh token: %w", err)
	}

	return p.GenerateTokens(ctx, claims.UserID)
}

// Generate token pair

func (p Provider) GenerateTokens(ctx context.Context, userID int64) (TokenPair, error) {
	accessToken, err := auth.GenerateAccessToken(userID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(userID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
	}, nil
}

// Logout

func (p Provider) Logout(ctx context.Context, refreshToken string) error {
	return nil
}

// Create project
func (p Provider) CreateProject(ctx context.Context, creatorID, projectID int64) error {
	exists, err := p.repository.UserExists(ctx, creatorID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	perm := models.ProjectPermission{
		UserID:    creatorID,
		ProjectID: projectID,
		Role:      RoleOwner,
	}

	err = p.repository.CreateProjectPermission(ctx, perm)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return ErrPermissionExists
		}
		if strings.Contains(err.Error(), "foreign key constraint") {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// Set permission
func (p Provider) AssignRole(ctx context.Context, userID, projectID int64, role string) error {
	if !isValidRole(role) {
		return fmt.Errorf("invalid role")
	}

	userExists, err := p.repository.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}
	if !userExists {
		return ErrUserNotFound
	}

	projectExists, err := p.repository.ProjectExists(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to check project: %w", err)
	}
	if !projectExists {
		return ErrProjectNotFound
	}

	currentRole, err := p.repository.GetUserProjectRole(ctx, userID, projectID)
	if err == nil && currentRole == RoleOwner && role != RoleOwner {
		return fmt.Errorf("cannot change owner role")
	}

	perm := models.ProjectPermission{
		UserID:    userID,
		ProjectID: projectID,
		Role:      role,
	}

	return p.repository.CreateProjectPermission(ctx, perm)
}

func isValidRole(role string) bool {
	return role == RoleOwner || role == RoleEditor || role == RoleViewer
}

// Get users role

func (p Provider) GetRole(ctx context.Context, userID, projectID int64) (string, error) {
	userExists, err := p.repository.UserExists(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return "", ErrUserNotFound
	}

	projectExists, err := p.repository.ProjectExists(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to check project existence: %w", err)
	}
	if !projectExists {
		return "", ErrProjectNotFound
	}

	role, err := p.repository.GetUserProjectRole(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrPermissionNotFound
		}
		return "", fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// UpdateRole
func (p Provider) UpdateRole(ctx context.Context, userID, projectID int64, role string) error {
	if !isValidRole(role) {
		return ErrInvalidRole
	}

	if exists, err := p.repository.UserExists(ctx, userID); err != nil || !exists {
		return ErrUserNotFound
	}
	if exists, err := p.repository.ProjectExists(ctx, projectID); err != nil || !exists {
		return ErrProjectNotFound
	}

	currentRole, err := p.repository.GetUserProjectRole(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("failed to get current role: %w", err)
	}

	if currentRole == RoleOwner && role != RoleOwner {
		return fmt.Errorf("cannot change owner role")
	}

	return p.repository.UpdateRole(ctx, userID, projectID, role)
}

func (p Provider) DeleteRole(ctx context.Context, userID, projectID int64) error {

	if exists, err := p.repository.UserExists(ctx, userID); err != nil || !exists {
		return ErrUserNotFound
	}
	if exists, err := p.repository.ProjectExists(ctx, projectID); err != nil || !exists {
		return ErrProjectNotFound
	}

	if role, err := p.repository.GetUserProjectRole(ctx, userID, projectID); err == nil && role == RoleOwner {
		return fmt.Errorf("cannot delete owner role")
	}

	return p.repository.DeletePermission(ctx, userID, projectID)
}

func (p Provider) DeleteUsersProjectPermission(ctx context.Context, userID, projectID int64) error {
	userExists, err := p.repository.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return ErrUserNotFound
	}

	projectExists, err := p.repository.ProjectExists(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to check project existence: %w", err)
	}
	if !projectExists {
		return ErrProjectNotFound
	}

	_, err = p.repository.GetUserProjectRole(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	return p.repository.DeletePermission(ctx, userID, projectID)
}
