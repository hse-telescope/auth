package storage

const (
	ChangeUsernameQuery = `
		UPDATE users
		SET username = $1
		WHERE email = $2 AND password = $3
	`
)
