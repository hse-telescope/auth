package storage

const (
	GetUserIDByUsername = `
		SELECT id
		FROM users
		WHERE username = $1
	`
)
