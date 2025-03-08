package users

import (
	"context"

	"github.com/hse-telescope/auth/internal/repository/models"
	"github.com/olegdayo/omniconv"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	AddUser(ctx context.Context, username, password string) (int64, error)
	CheckUser(ctx context.Context, username string) (int64, string, error)
}

type Provider struct {
	repository Repository
}

func New(repository Repository) Provider {
	return Provider{repository: repository}
}

// GetAllUsers

func (p Provider) GetUsers(ctx context.Context) ([]User, error) {
	users, err := p.repository.GetUsers(ctx)

	if err != nil {
		return nil, err
	}

	return omniconv.ConvertSlice(users, DBUser2ProviderUser), nil
}

// Register

func (p Provider) AddUser(ctx context.Context, username, password string) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return -1, err
	}

	userID, err := p.repository.AddUser(ctx, username, string(hashedPassword))
	if err != nil {
		return -1, err
	}

	return userID, nil
}

// Login

func (p Provider) LoginUser(ctx context.Context, username, password string) (int64, error) {
	userID, hashedPassword, err := p.repository.CheckUser(ctx, username)
	if err != nil {
		return -1, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return -1, err
	}

	return userID, nil
}
