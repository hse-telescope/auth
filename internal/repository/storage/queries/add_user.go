package storage

const (
	AddUserQuery = `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id
	`
)
