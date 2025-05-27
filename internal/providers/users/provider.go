package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hse-telescope/auth/internal/auth"
	"github.com/hse-telescope/auth/internal/repository/models"
	"github.com/hse-telescope/auth/internal/repository/storage"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	AddUser(ctx context.Context, username, email, password string) (int64, error)
	CheckUserByUsername(ctx context.Context, username string) (int64, string, error)
	CheckUserByEmail(ctx context.Context, email string) (int64, string, error)

	GetUserProjects(ctx context.Context, userID int64) ([]int64, error)
	GetProjectUsers(ctx context.Context, projectID int64) ([]models.ProjectUser, error)
	CreateProjectPermission(ctx context.Context, perm models.ProjectPermission) error

	AssignRole(ctx context.Context, perm models.ProjectPermission) error
	GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error)
	UpdateRole(ctx context.Context, userID, projectID int64, role string) error
	DeletePermission(ctx context.Context, userID, projectID int64) error

	UserExists(ctx context.Context, userID int64) (bool, error)
	ProjectExists(ctx context.Context, projectID int64) (bool, error)
	GetUserIDByUsername(ctx context.Context, username string) (int64, error)
	GetUserIDByEmail(ctx context.Context, email string) (int64, error)

	ChangeUsername(ctx context.Context, username, email, password string) error
	ChangeEmail(ctx context.Context, username, email, password string) error
	ChangePassword(ctx context.Context, username, email, password string) error
	GetUsernameByEmail(ctx context.Context, email string) (string, error)
}

type Provider struct {
	repository Repository
}

func New(repository Repository) Provider {
	return Provider{repository: repository}
}

var (
	ErrUsernameExists         = errors.New("username already exists")
	ErrEmailExists            = errors.New("email already exists")
	ErrIncorrectPassword      = errors.New("incorrect password")
	ErrPermissionExists       = errors.New("permission already exists")
	ErrInvalidRole            = errors.New("invalid role")
	ErrUserNotFound           = errors.New("user not found")
	ErrProjectNotFound        = errors.New("project not found")
	ErrPermissionNotFound     = errors.New("permission not found")
	ErrProjectExists          = errors.New("project already exists")
	ErrAssignerNotFound       = errors.New("assigner not found")
	ErrAssignableNotFound     = errors.New("assignable not found")
	ErrAssignerRoleNotFound   = errors.New("assigner role not found")
	ErrAssignableRoleNotFound = errors.New("assignable role not found")
	ErrAssignerIsNotOwner     = errors.New("assigner is not owner")
	ErrOwnerRoleChanging      = errors.New("cannot change owner role")
	ErrPasswordConflict       = errors.New("passwords by username and email are different")
	ErrGeneratePassword       = errors.New("password generate error")
	ErrHashingPassword        = errors.New("password hashing error")
)

//////////////////
//				//
//	Auth base	//
//				//
//////////////////

// Register
func (p Provider) RegisterUser(ctx context.Context, username, email, password string) (TokenPair, error) {
	_, _, err := p.repository.CheckUserByUsername(ctx, username)
	if err == nil {
		return TokenPair{}, ErrUsernameExists
	} else if err != sql.ErrNoRows {
		return TokenPair{}, fmt.Errorf("failed to check user: %w", err)
	}

	_, _, err = p.repository.CheckUserByEmail(ctx, email)
	if err == nil {
		return TokenPair{}, ErrEmailExists
	} else if err != sql.ErrNoRows {
		return TokenPair{}, fmt.Errorf("failed to check user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := p.repository.AddUser(ctx, username, email, string(hashedPassword))
	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to add user: %w", err)
	}

	return p.GenerateTokens(ctx, userID)
}

// Login
func (p Provider) LoginUser(ctx context.Context, loginData, password string) (TokenPair, error) {
	var userID int64
	var hashedPassword string
	var err error

	if strings.Contains(loginData, "@") {
		userID, hashedPassword, err = p.repository.CheckUserByEmail(ctx, loginData)
	} else {
		userID, hashedPassword, err = p.repository.CheckUserByUsername(ctx, loginData)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return TokenPair{}, ErrUserNotFound
		}
		fmt.Println(err.Error())
		return TokenPair{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		fmt.Println(err.Error())
		return TokenPair{}, ErrIncorrectPassword
	}

	return p.GenerateTokens(ctx, userID)
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

//////////////////////
//					//
//	Role managment	//
//					//
//////////////////////

// Get user projects
func (p Provider) GetUserProjects(ctx context.Context, userID int64) ([]int64, error) {
	exists, err := p.repository.UserExists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	projectIDs, err := p.repository.GetUserProjects(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projectIDs, nil
}

// Create project
func (p Provider) CreateProject(ctx context.Context, creatorID, projectID int64) error {
	userExists, err := p.repository.UserExists(ctx, creatorID)
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
	if projectExists {
		return ErrProjectExists
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

// Set permission
func (p Provider) AssignRole(ctx context.Context, userID int64, username string, projectID int64, role string) error {
	if !isValidRole(role) {
		return ErrInvalidRole
	}

	userExists, err := p.repository.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check owner: %w", err)
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

	assignerRole, err := p.repository.GetUserProjectRole(ctx, userID, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAssignerRoleNotFound
		}
		return fmt.Errorf("failed to check ownership: %w", err)
	}
	if assignerRole != "owner" {
		return ErrAssignerIsNotOwner
	}

	assignableUserID, err := p.repository.GetUserIDByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAssignableNotFound
		}
		return fmt.Errorf("failed to get user_id from username: %w", err)
	}

	currentRole, err := p.repository.GetUserProjectRole(ctx, assignableUserID, projectID)
	if err != nil && err != sql.ErrNoRows {
		if err == storage.ErrPermissionExists {
			return ErrPermissionExists
		}
		return fmt.Errorf("failed to check assignable role: %w", err)
	}
	if currentRole == RoleOwner {
		return ErrOwnerRoleChanging
	}

	perm := models.ProjectPermission{
		UserID:    assignableUserID,
		ProjectID: projectID,
		Role:      role,
	}

	err = p.repository.AssignRole(ctx, perm)
	if err == storage.ErrPermissionExists {
		return ErrPermissionExists
	}

	return nil
}

// UpdateRole
func (p Provider) UpdateRole(ctx context.Context, userID int64, username string, projectID int64, role string) error {
	if !isValidRole(role) {
		return ErrInvalidRole
	}

	userExists, err := p.repository.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check owner: %w", err)
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

	assignerRole, err := p.repository.GetUserProjectRole(ctx, userID, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAssignerNotFound
		}
		return fmt.Errorf("failed to check ownership: %w", err)
	}
	if assignerRole != "owner" {
		return ErrAssignerIsNotOwner
	}

	assignableUserID, err := p.repository.GetUserIDByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAssignableNotFound
		}
		return fmt.Errorf("failed to get user_id from username: %w", err)
	}

	currentRole, err := p.repository.GetUserProjectRole(ctx, assignableUserID, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAssignableRoleNotFound
		}
		return fmt.Errorf("failed to check assignable role: %w", err)
	}
	if currentRole == RoleOwner {
		return ErrOwnerRoleChanging
	}

	return p.repository.UpdateRole(ctx, assignableUserID, projectID, role)
}

func (p Provider) DeleteRole(ctx context.Context, userID int64, username string, projectID int64) error {
	userExists, err := p.repository.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check owner: %w", err)
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

	assignerRole, err := p.repository.GetUserProjectRole(ctx, userID, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAssignerNotFound
		}
		return fmt.Errorf("failed to check ownership: %w", err)
	}
	if assignerRole != "owner" {
		return ErrAssignerIsNotOwner
	}

	assignableUserID, err := p.repository.GetUserIDByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAssignableNotFound
		}
		return fmt.Errorf("failed to get user_id from username: %w", err)
	}

	currentRole, err := p.repository.GetUserProjectRole(ctx, assignableUserID, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAssignableRoleNotFound
		}
		return fmt.Errorf("failed to check assignable role: %w", err)
	}
	if currentRole == RoleOwner {
		return ErrOwnerRoleChanging
	}

	return p.repository.DeletePermission(ctx, assignableUserID, projectID)
}

func isValidRole(role string) bool {
	return (role != RoleOwner) && (role == RoleEditor || role == RoleViewer)
}

func (p Provider) GetProjectUserRoles(ctx context.Context, projectID int64) ([]models.ProjectUser, error) {
	users, err := p.repository.GetProjectUsers(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return users, nil
}

///////////////////////
//					 //
// Change login data //
//					 //
///////////////////////

func (p Provider) ChangeUsername(ctx context.Context, oldUsername, newUsername, email, password string) error {
	_, hashedPassword, err := p.repository.CheckUserByUsername(ctx, oldUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return ErrIncorrectPassword
	}

	err = p.repository.ChangeUsername(ctx, newUsername, email, hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		} else if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint ") {
			return ErrUsernameExists
		}
		return err
	}

	return nil
}

func (p Provider) ChangeEmail(ctx context.Context, username, oldEmail, newEmail, password string) error {
	_, hashedPassword, err := p.repository.CheckUserByEmail(ctx, oldEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return ErrIncorrectPassword
	}

	err = p.repository.ChangeEmail(ctx, username, newEmail, hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		} else if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint ") {
			return ErrEmailExists
		}
		return err
	}

	return nil
}

func (p Provider) ChangePassword(ctx context.Context, username, email, currPassword, newPassword string) error {
	_, hashedPasswordByEmail, err := p.repository.CheckUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}
	_, hashedPasswordByUsername, err := p.repository.CheckUserByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	log.Default().Printf("\n[HASHED BY USERNAME:] %s\n[HASHED BY EMAIL:   ] %s", hashedPasswordByUsername, hashedPasswordByEmail)

	if hashedPasswordByEmail != hashedPasswordByUsername {
		return ErrPasswordConflict
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPasswordByUsername), []byte(currPassword))
	if err != nil {
		log.Default().Printf("[OLD:] %s \n [NEW:] %s", hashedPasswordByUsername, currPassword)
		return ErrIncorrectPassword
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return ErrHashingPassword
	}

	err = p.repository.ChangePassword(ctx, username, email, string(hashedNewPassword))
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (p Provider) ForgotPassword(ctx context.Context, email string) error {
	newPassword, err := auth.GenerateNewPassword()
	if err != nil {
		return ErrGeneratePassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return ErrHashingPassword
	}

	username, err := p.repository.GetUsernameByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	// err := SendPassword... Kafka

	err = p.repository.ChangePassword(ctx, username, email, string(hashedPassword))
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	} else {
		log.Default().Printf("New password: %s", newPassword)
	}

	return nil
}
