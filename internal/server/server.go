package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hse-telescope/auth/internal/config"
	"github.com/hse-telescope/auth/internal/providers/users"
	"github.com/hse-telescope/auth/internal/repository/models"
	"github.com/rs/cors"
)

type Provider interface {
	RegisterUser(ctx context.Context, username, email, password string) (users.TokenPair, error)
	LoginUser(ctx context.Context, login, password string) (users.TokenPair, error)
	//RefreshTokens(ctx context.Context, refreshToken string) (users.TokenPair, error)
	GenerateTokens(ctx context.Context, userID int64) (users.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error

	GetUserProjects(ctx context.Context, userID int64) ([]int64, error)
	GetProjectUserRoles(ctx context.Context, projectID int64) ([]models.ProjectUser, error)

	CreateProject(ctx context.Context, userID, projectID int64) error
	GetRole(ctx context.Context, userID, projectID int64) (string, error)
	AssignRole(ctx context.Context, userID int64, username string, projectID int64, role string) error
	UpdateRole(ctx context.Context, userID int64, username string, projectID int64, role string) error
	DeleteRole(ctx context.Context, userID int64, username string, projectID int64) error

	ChangeUsername(ctx context.Context, oldUsername, newUsername, email, password string) error
	ChangeEmail(ctx context.Context, username, oldEmail, newEmail, password string) error
	ChangePassword(ctx context.Context, username, email, oldPassword, newPassword string) error
}

type Server struct {
	server   http.Server
	provider Provider
}

func New(conf config.Config, provider Provider) *Server {
	s := new(Server)
	s.server.Addr = fmt.Sprintf(":%d", conf.Port)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	s.server.Handler = c.Handler(s.setRouter())
	s.provider = provider

	return s
}

func (s *Server) setRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("OPTIONS /{path...}", func(w http.ResponseWriter, r *http.Request) {
		setCommonHeaders(w)
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /ping", s.pingHandler)

	mux.HandleFunc("POST /register", s.registerUserHandler)
	mux.HandleFunc("POST /login", s.loginUserHandler)
	//mux.HandleFunc("POST /refresh", s.refreshHandler)
	mux.HandleFunc("POST /logout", s.logoutHandler)

	mux.HandleFunc("GET /usersProjects", s.getUserProjectsHandler)
	mux.Handle("GET /projectUsers", s.AuthMiddleware(http.HandlerFunc(s.getProjectUsersHandler)))

	mux.HandleFunc("POST /createProject", s.createProjectHandler)

	mux.HandleFunc("GET /userProjectRole", s.getUserProjectRoleHandler)
	mux.Handle("POST /assignRole", s.AuthMiddleware(http.HandlerFunc(s.assignRoleHandler)))
	mux.Handle("PUT /updateRole", s.AuthMiddleware(http.HandlerFunc(s.updateRoleHandler)))
	mux.Handle("DELETE /deleteRole", s.AuthMiddleware(http.HandlerFunc(s.deleteRoleHandler)))

	//mux.HandleFunc("POST /forgotPassword", s.forgotPasswordHandler)

	mux.HandleFunc("PUT /username", s.changeUsernameHandler)
	mux.HandleFunc("PUT /email", s.changeEmailHandler)
	mux.HandleFunc("PUT /password", s.changePasswordHandler)

	return mux
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}
