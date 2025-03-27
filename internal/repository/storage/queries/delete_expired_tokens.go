package storage

const (
	DeleteExpiredRefreshTokensQuery = `
		DELETE FROM refresh_tokens WHERE expires_at < NOW()
	`
)
