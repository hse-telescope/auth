package server

import (
	"context"
	"encoding/json"
	"net/http"
)

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.provider.GetUsers(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong: " + err.Error()))
		return
	}
	body, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
