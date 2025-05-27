package models

type ProjectPermission struct {
	UserID    int64  `db:"user_id"`
	ProjectID int64  `db:"project_id"`
	Role      string `db:"role"`
}

type ProjectUser struct {
	Username string `db:"username" json:"username"`
	Role     string `db:"role" json:"role"`
}
