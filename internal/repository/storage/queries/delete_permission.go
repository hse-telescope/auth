package storage

const (
	DeletePermissionQuery = `
		DELETE FROM project_permissions WHERE user_id = $1 AND project_id = $2
	`
)
