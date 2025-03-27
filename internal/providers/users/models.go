package users

import (
	"time"

	"github.com/hse-telescope/auth/internal/repository/models"
)

type User struct {
	ID       int64
	Username string
	Password string
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	UserID       int64
}

type ExpiredTokenError struct {
	ExpiredAt time.Time
	Now       time.Time
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

func DBToken2ProviderToken(refreshToken models.RefreshToken) RefreshToken {
	return RefreshToken{
		ID:        refreshToken.ID,
		UserID:    refreshToken.UserID,
		Token:     refreshToken.Token,
		ExpiresAt: refreshToken.ExpiresAt,
		CreatedAt: refreshToken.CreatedAt,
	}
}
