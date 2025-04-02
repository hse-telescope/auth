package server

import (
	"github.com/hse-telescope/auth/internal/providers/users"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
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
