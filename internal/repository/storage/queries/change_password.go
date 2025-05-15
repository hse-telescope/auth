package storage

const (
	ChangePasswordQuery = `
		UPDATE users
		SET password = $3
		WHERE username = $1 AND email = $2
	`
)
