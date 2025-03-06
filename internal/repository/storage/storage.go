package storage

import (
	"context"
	"database/sql"

	"github.com/hse-telescope/auth/internal/repository/models"
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
	q := `SELECT * FROM people`

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

// LoginUser

// DeleteUser

// ChangePassword

// ChangeUsername
