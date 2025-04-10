package storage

import (
	"context"
	"database/sql"
	"errors"
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

func (s Storage) CreateProjectPermission(ctx context.Context, perm models.ProjectPermission) error {
	res, err := s.db.ExecContext(ctx, storage.SetPermissionQuery,
		perm.UserID,
		perm.ProjectID,
		perm.Role,
	)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("permission already exists")
	}

	return nil
}

func (s Storage) GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error) {
	q := storage.GetPermissionQuery
	var role string
	err := s.db.QueryRowContext(ctx, q, userID, projectID).Scan(&role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	return role, err
}

func (s Storage) UserExists(ctx context.Context, userID int64) (bool, error) {
	q := storage.UserExistsQuery
	var exists bool
	err := s.db.QueryRowContext(ctx, q, userID).Scan(&exists)
	return exists, err
}

func (s Storage) ProjectExists(ctx context.Context, projectID int64) (bool, error) {
	q := storage.ProjectExistsQuery
	var exists bool
	err := s.db.QueryRowContext(ctx, q, projectID).Scan(&exists)
	return exists, err
}

func (s Storage) UpdateRole(ctx context.Context, userID, projectID int64, role string) error {
	res, err := s.db.ExecContext(ctx, storage.UpdatePermissionQuery,
		userID, projectID, role,
	)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s Storage) DeletePermission(ctx context.Context, userID, projectID int64) error {
	q := storage.DeletePermissionQuery
	_, err := s.db.ExecContext(ctx, q, userID, projectID)
	return err
}
