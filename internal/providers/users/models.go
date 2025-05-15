package users

const (
	RoleOwner  = "owner"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

type User struct {
	ID       int64
	Username string
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	UserID       int64
}
