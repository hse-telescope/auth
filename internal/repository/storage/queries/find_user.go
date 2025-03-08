package storage

const (
	FindUserQuery = `
		SELECT id, password
		FROM users
		WHERE username = $1
	`
)
