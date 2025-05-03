package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hse-telescope/auth/internal/config"
	"github.com/hse-telescope/auth/internal/providers/users"
	"github.com/rs/cors"
)

type Provider interface {
	RegisterUser(ctx context.Context, username, email, password string) (users.TokenPair, error)
	LoginUser(ctx context.Context, login, password string) (users.TokenPair, error)
	//RefreshTokens(ctx context.Context, refreshToken string) (users.TokenPair, error)
	GenerateTokens(ctx context.Context, userID int64) (users.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error

	GetUserProjects(ctx context.Context, userID int64) ([]int64, error)
	CreateProject(ctx context.Context, userID, projectID int64) error
	GetRole(ctx context.Context, userID, projectID int64) (string, error)
	AssignRole(ctx context.Context, userID int64, username string, projectID int64, role string) error
	UpdateRole(ctx context.Context, userID int64, username string, projectID int64, role string) error
	DeleteRole(ctx context.Context, userID int64, username string, projectID int64) error
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
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
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

	// "POST /"

	mux.HandleFunc("GET /usersProjects", s.getUserProjectsHandler)

	mux.HandleFunc("POST /createProject", s.createProjectHandler)

	mux.HandleFunc("GET /userProjectRole", s.getUserProjectRoleHandler)
	mux.HandleFunc("POST /assignRole", s.assignRoleHandler)
	mux.HandleFunc("PUT /updateRole", s.updateRoleHandler)
	mux.HandleFunc("DELETE /deleteRole", s.deleteRoleHandler)

	return mux
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}
