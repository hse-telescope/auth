package storage

const (
	DeleteRefreshTokenQuery = `
		DELETE FROM refresh_tokens WHERE token = $1
	`
)
