package server

import (
	"github.com/hse-telescope/auth/internal/providers/users"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RoleAssignmentRequest struct {
	UserID    int64  `json:"user_id"`
	ProjectID int64  `json:"project_id"`
	Role      string `json:"role"`
}

type ProjectRoleRequest struct {
	UserID    int64 `json:"user_id"`
	ProjectID int64 `json:"project_id"`
}

type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenRequest struct {
	RefreshToken string `json:"token"`
}

func ServerUser2ProviderUser(user User) users.User {
	return users.User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
	}
}

func ProviderUser2ServerUser(user users.User) User {
	return User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
	}
}
