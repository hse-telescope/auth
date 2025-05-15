package storage

const (
	UpdatePermissionQuery = `
        UPDATE project_permissions 
        SET role = $3 
        WHERE user_id = $1 AND project_id = $2
    `
)
