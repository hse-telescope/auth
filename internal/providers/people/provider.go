package users

import (
	"context"

	"github.com/hse-telescope/auth/internal/repository/models"
	"github.com/olegdayo/omniconv"
)

type Repository interface {
	GetUsers(ctx context.Context) (models.User, error)
}

type Provider struct {
	repository Repository
}

func New(repository Repository) Provider {
	return Provider{repository: repository}
}

//TO DO

// GetAllUsers

func (p Provider) GetUsers(ctx context.Context) ([]User, error) {
	users, err := p.repository.GetUsers(ctx)

	if err != nil {
		return nil, err
	}

	return omniconv.ConvertSlice(users, DBUser2ProviderUser), nil
}

// Register

// Login
