package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hse-telescope/auth/internal/auth"
	"github.com/hse-telescope/auth/internal/repository/models"
	"github.com/olegdayo/omniconv"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	AddUser(ctx context.Context, username, password string) (int64, error)
	CheckUser(ctx context.Context, username string) (int64, string, error)
	GetRefreshToken(ctx context.Context, refreshToken string) (models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, refreshToken string) error
	CreateRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
}

type Provider struct {
	repository Repository
}

func (e ExpiredTokenError) Error() string {
	return fmt.Sprintf("refresh token expired at %s (current time: %s)",
		e.ExpiredAt.Format(time.RFC1123),
		e.Now.Format(time.RFC1123))
}

func New(repository Repository) Provider {
	return Provider{repository: repository}
}

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

func (p Provider) RefreshTokens(ctx context.Context, refreshToken string) (TokenPair, error) {
	token, err := p.repository.GetRefreshToken(ctx, refreshToken)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TokenPair{}, fmt.Errorf("invalid refresh token")
		}
		return TokenPair{}, fmt.Errorf("failed to get refresh token: %w", err)
	}

	if time.Now().After(token.ExpiresAt) {
		if err := p.repository.DeleteRefreshToken(ctx, refreshToken); err != nil {
			return TokenPair{}, fmt.Errorf("expired refresh token deleting error: %w", err)
		}
		return TokenPair{}, ExpiredTokenError{
			ExpiredAt: token.ExpiresAt,
			Now:       time.Now(),
		}
	}

	if err := p.repository.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return TokenPair{}, fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return p.GenerateTokens(ctx, token.UserID)
}

func (p Provider) GenerateTokens(ctx context.Context, userID int64) (TokenPair, error) {
	accessToken, err := auth.GenerateAccessToken(userID)

	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, expires_at, err := auth.GenerateRefreshToken()

	if err != nil {
		return TokenPair{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := p.repository.CreateRefreshToken(ctx, userID, refreshToken, expires_at); err != nil {
		return TokenPair{}, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken, UserID: userID}, nil
}

func (p Provider) Logout(ctx context.Context, refreshToken string) error {
	return p.repository.DeleteRefreshToken(ctx, refreshToken)
}
