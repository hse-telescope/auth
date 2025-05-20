package storage

const (
	GetUsernameByEmailQuery = `
		SELECT username
		FROM users
		WHERE email = $1
	`
)
