package users

import (
	"github.com/hse-telescope/auth/internal/repository/models"
)

const (
	RoleOwner  = "owner"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

type User struct {
	ID       int64
	Username string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	UserID       int64
}

func ProviderUser2DBUser(user User) models.User {
	return models.User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
	}
}

func DBUser2ProviderUser(user models.User) User {
	return User{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
	}
}
