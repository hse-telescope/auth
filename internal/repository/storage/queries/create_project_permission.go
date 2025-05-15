package storage

const (
	CreateProjectPermissionQuery = `
    	INSERT INTO project_permissions (user_id, project_id, role)
    	SELECT $1, $2, $3
    	WHERE NOT EXISTS (
			SELECT 1 FROM project_permissions
			WHERE project_id = $2 AND role = 'owner'
		)
    	RETURNING user_id
    `
)
