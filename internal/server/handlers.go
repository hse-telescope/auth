package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hse-telescope/auth/internal/providers/users"
)

type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenRequest struct {
	RefreshToken string `json:"token"`
}

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Ping
func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)
	w.Write([]byte("pong"))
}

// Get all users
func (s *Server) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)
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

// Login user
func (s *Server) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	var req CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := s.provider.LoginUser(context.Background(), req.Username, req.Password)
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
		"message":       "Login successful!",
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Register new user
func (s *Server) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	var req CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := s.provider.RegisterUser(context.Background(), req.Username, req.Password)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already registered", http.StatusConflict)
		} else {
			http.Error(w, "Could not register user", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"message":       "User registered!",
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Logout

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	var req TokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := s.provider.Logout(context.Background(), req.RefreshToken)

	if err != nil {
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Logout successful!",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Refresh

func (s *Server) refreshHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	var req TokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := s.provider.RefreshTokens(context.Background(), req.RefreshToken)

	if err != nil {
		var expiredErr users.ExpiredTokenError
		switch {
		case errors.Is(err, sql.ErrNoRows) || err.Error() == "invalid refresh token":
			s.respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		case errors.As(err, &expiredErr):
			s.respondWithError(w, http.StatusUnauthorized,
				fmt.Sprintf("Token expired at %s , compared with %s",
					expiredErr.ExpiredAt.Format(time.RFC1123), expiredErr.Now.Format(time.RFC1123)))
		default:
			s.respondWithError(w, http.StatusInternalServerError,
				fmt.Sprintf("Refresh failed: %v", err))
		}
		return
	}

	response := map[string]interface{}{
		"message":       "Refresh successful!",
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) respondWithError(w http.ResponseWriter, code int, message string) {
	s.respondWithJSON(w, code, map[string]string{"error": message})
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	setCommonHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to encode response"))
	}
}
