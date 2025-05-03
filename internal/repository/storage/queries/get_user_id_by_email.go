package storage

const (
	GetUserIDByEmail = `
		SELECT id
		FROM users
		WHERE email = $1
	`
)
