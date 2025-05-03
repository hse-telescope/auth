package server

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GetUserProjectsRequest struct {
	UserID int64 `json:"user_id"`
}

type GetRoleRequest struct {
	UserID    int64  `json:"user_id"`
	ProjectID int64  `json:"project_id"`
	Role      string `json:"role"`
}

type CreateProjectPermissionRequest struct {
	UserID    int64 `json:"user_id"`
	ProjectID int64 `json:"project_id"`
}

type AssignRoleRequest struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	ProjectID int64  `json:"project_id"`
	Role      string `json:"role"`
}

type UpdateRoleRequest struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	ProjectID int64  `json:"project_id"`
	Role      string `json:"role"`
}

type DeleteRoleRequest struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	ProjectID int64  `json:"project_id"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	LoginData string `json:"loginData"`
	Password  string `json:"password"`
}

type TokenRequest struct {
	RefreshToken string `json:"token"`
}
