package storage

const (
	ChangeEmailQuery = `
		UPDATE users
		SET email = $2
		WHERE username = $1 AND password = $3
	`
)
