package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hse-telescope/auth/internal/repository/models"
	storage "github.com/hse-telescope/auth/internal/repository/storage/queries"
	"github.com/hse-telescope/utils/db/psql"
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

var (
	ErrPermissionExists = errors.New("permission already exists [storage]")
	ErrOwnerExists      = errors.New("owner already exists")
)

func (s Storage) AddUser(ctx context.Context, username, email, hashedPassword string) (int64, error) {
	q := storage.AddUserQuery
	var userID int64
	err := s.db.QueryRowContext(ctx, q, username, email, hashedPassword).Scan(&userID)

	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	return userID, nil
}

func (s Storage) CheckUserByUsername(ctx context.Context, username string) (int64, string, error) {
	q := storage.FindUserByUsernameQuery
	var userID int64
	var hashedPassword string
	err := s.db.QueryRowContext(ctx, q, username).Scan(&userID, &hashedPassword)
	if err != nil {
		return -1, "", err
	}
	return userID, hashedPassword, err
}

func (s Storage) CheckUserByEmail(ctx context.Context, email string) (int64, string, error) {
	q := storage.FindUserByEmailQuery
	var userID int64
	var hashedPassword string
	err := s.db.QueryRowContext(ctx, q, email).Scan(&userID, &hashedPassword)
	if err != nil {
		return -1, "", err
	}
	return userID, hashedPassword, err
}

func (s Storage) CreateProjectPermission(ctx context.Context, perm models.ProjectPermission) error {
	res, err := s.db.ExecContext(ctx, storage.CreateProjectPermissionQuery,
		perm.UserID,
		perm.ProjectID,
		perm.Role,
	)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrOwnerExists
	}

	return nil
}

func (s Storage) AssignRole(ctx context.Context, perm models.ProjectPermission) error {
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
		return ErrPermissionExists
	}

	return nil
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

func (s Storage) GetUserProjects(ctx context.Context, userID int64) ([]int64, error) {
	q := storage.GetUserProjects
	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		projectIDs = append(projectIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projectIDs, nil
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

func (s Storage) GetUserIDByUsername(ctx context.Context, username string) (int64, error) {
	q := storage.GetUserIDByUsername
	var userID int64
	err := s.db.QueryRowContext(ctx, q, username).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, sql.ErrNoRows
		}
		return -1, err
	}
	return userID, nil
}

func (s Storage) GetUserIDByEmail(ctx context.Context, email string) (int64, error) {
	q := storage.GetUserIDByEmail
	var userID int64
	err := s.db.QueryRowContext(ctx, q, email).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, sql.ErrNoRows
		}
		return -1, err
	}
	return userID, nil
}

func (s Storage) ChangeUsername(ctx context.Context, username, email, password string) error {
	q := storage.ChangeUsernameQuery
	res, err := s.db.ExecContext(ctx, q, username, email, password)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s Storage) ChangeEmail(ctx context.Context, username, email, password string) error {
	q := storage.ChangeEmailQuery
	res, err := s.db.ExecContext(ctx, q, username, email, password)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s Storage) ChangePassword(ctx context.Context, username, email, password string) error {
	q := storage.ChangePasswordQuery
	res, err := s.db.ExecContext(ctx, q, username, email, password)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
