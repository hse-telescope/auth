package users

import "github.com/hse-telescope/auth/internal/repository/models"

type User struct {
	ID       int
	Username string
	Password string
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
