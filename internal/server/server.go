package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hse-telescope/auth/internal/config"
	"github.com/hse-telescope/auth/internal/providers/users"
)

type Provider interface {
	GetUsers(ctx context.Context) ([]users.User, error)
}

type Server struct {
	server   http.Server
	provider Provider
}

func New(conf config.Config, provider Provider) *Server {
	s := new(Server)
	s.server.Addr = fmt.Sprintf(":%d", conf.Port)
	s.server.Handler = s.setRouter()
	s.provider = provider
	return s
}

func (s *Server) setRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", s.pingHandler)
	mux.HandleFunc("GET /users", s.getUsersHandler)
	return mux
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}
