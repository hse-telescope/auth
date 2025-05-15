package storage

const (
	FindUserByUsernameQuery = `
        SELECT id, password
        FROM users
        WHERE username = $1
    `

	FindUserByEmailQuery = `
        SELECT id, password
        FROM users
        WHERE email = $1
    `
)
