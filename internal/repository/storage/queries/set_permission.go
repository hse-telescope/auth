package storage

const (
	SetPermissionQuery = `
        INSERT INTO project_permissions 
        (user_id, project_id, role)
        VALUES ($1, $2, $3)
        ON CONFLICT DO NOTHING
    `
)
