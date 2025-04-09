package storage

const (
	GetPermissionQuery = `
		SELECT role 
		FROM project_permissions 
		WHERE user_id = $1 AND project_id = $2
	`
)
