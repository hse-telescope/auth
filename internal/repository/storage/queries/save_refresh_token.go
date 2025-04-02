package storage

const (
	SaveRefreshTokenQuery = `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET token = $2, expires_at = $3, created_at = NOW()
	`
)
