package server

import (
	"context"
	"encoding/json"
	"net/http"
)

type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Ping
func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Write([]byte("pong"))
}

// Get all users
func (s *Server) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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

// Register new user
func (s *Server) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var req CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	registeredUserID, err := s.provider.AddUser(context.Background(), req.Username, req.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already registered", http.StatusConflict)
		} else {
			http.Error(w, "Could not register user", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"message": "User registered!",
		"id":      registeredUserID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Login user
func (s *Server) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var req CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	loginUserID, err := s.provider.LoginUser(context.Background(), req.Username, req.Password)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User not found", http.StatusNotFound)
		} else if err.Error() == "incorrect password" {
			http.Error(w, "Incorrect password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Could not login user", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"message": "Login successful!",
		"id":      loginUserID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
