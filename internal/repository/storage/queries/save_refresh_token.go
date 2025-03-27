package storage

const (
	SaveRefreshTokenQuery = `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id)
		DO UPDATE SET token = $2, expires_at = $3
	`
)
