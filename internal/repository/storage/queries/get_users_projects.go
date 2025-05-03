package storage

const (
	GetUserProjects = `
		SELECT project_id
		FROM project_permissions
		WHERE user_id = $1
	`
)
