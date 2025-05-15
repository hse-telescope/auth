package storage

const (
	ProjectExistsQuery = `
		SELECT EXISTS(SELECT 1 FROM project_permissions WHERE project_id = $1)
	`
)
