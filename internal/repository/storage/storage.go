package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hse-telescope/auth/internal/repository/models"
	storage "github.com/hse-telescope/auth/internal/repository/storage/queries"
	"github.com/hse-telescope/utils/db/psql"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sql.DB
}

func New(dbURL string, migrationsPath string) (Storage, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return Storage{}, err
	}
	err = db.Ping()
	if err != nil {
		return Storage{}, err
	}
	psql.MigrateDB(db, migrationsPath, psql.PGDriver)
	return Storage{
		db: db,
	}, nil
}

//TO DO

// GetAllUsers

func (s Storage) GetUsers(ctx context.Context) ([]models.User, error) {
	q := storage.GetUsersQuery

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0)

	err = sqlx.StructScan(rows, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// RegisterUser

func (s Storage) AddUser(ctx context.Context, username, hashedPassword string) (int64, error) {
	q := storage.AddUserQuery
	var userID int64
	err := s.db.QueryRowContext(ctx, q, username, hashedPassword).Scan(&userID)

	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	return userID, nil
}

// LoginUser

func (s Storage) CheckUser(ctx context.Context, username string) (int64, string, error) {
	q := storage.FindUserQuery
	var userID int64
	var hashedPassword string
	err := s.db.QueryRowContext(ctx, q, username).Scan(&userID, &hashedPassword)
	if err != nil {
		fmt.Println(err.Error())
		return -1, "", err
	}
	return userID, hashedPassword, nil
}

// DeleteUser

// ChangePassword

// ChangeUsername
