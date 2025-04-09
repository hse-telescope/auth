package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hse-telescope/auth/internal/providers/users"
)

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

	response := map[string]string{
		"message": "Logout successful (note: refresh tokens are still valid)",
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
		if err.Error() == "invalid refresh token" {
			s.respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		} else {
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

// Create project

func (s *Server) createProjectHandler(w http.ResponseWriter, r *http.Request) {

	var req ProjectRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}

	err := s.provider.CreateProject(r.Context(), req.UserID, req.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrPermissionExists):
			s.respondWithError(w, http.StatusConflict, "project already exists")
		default:
			s.respondWithError(w, http.StatusInternalServerError, "failed to create project")
		}
		return
	}

	s.respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"project_id": req.ProjectID,
		"message":    "project created successfully",
	})

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// Get users role in project

func (s *Server) getUserProjectRoleHandler(w http.ResponseWriter, r *http.Request) {

	var req ProjectRoleRequest
	var err error

	req.UserID, err = strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)

	if err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request: user_id parse error")
		return
	}
	req.ProjectID, err = strconv.ParseInt(r.URL.Query().Get("project_id"), 10, 64)

	if err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request: project_id parse error")
		return
	}

	role, err := s.provider.GetRole(r.Context(), req.UserID, req.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrProjectNotFound):
			s.respondWithError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, users.ErrPermissionNotFound):
			s.respondWithError(w, http.StatusNotFound, "permission not found")
		default:
			log.Printf("GetRole error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to get role")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"role": role})
}

// Assign role

func (s *Server) assignRoleHandler(w http.ResponseWriter, r *http.Request) {

	var req RoleAssignmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}

	err := s.provider.AssignRole(r.Context(), req.UserID, req.ProjectID, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrProjectNotFound):
			s.respondWithError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, users.ErrInvalidRole):
			s.respondWithError(w, http.StatusBadRequest, "invalid role")
		case errors.Is(err, users.ErrPermissionExists):
			s.respondWithError(w, http.StatusConflict, "role already assigned")
		default:
			log.Printf("AssignRole error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to assign role")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// Update role
func (s *Server) updateRoleHandler(w http.ResponseWriter, r *http.Request) {
	var req RoleAssignmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}

	err := s.provider.UpdateRole(r.Context(), req.UserID, req.ProjectID, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrProjectNotFound):
			s.respondWithError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, users.ErrPermissionNotFound):
			s.respondWithError(w, http.StatusNotFound, "permission not found")
		case errors.Is(err, users.ErrInvalidRole):
			s.respondWithError(w, http.StatusBadRequest, "invalid role")
		case strings.Contains(err.Error(), "cannot change owner"):
			s.respondWithError(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("UpdateRole error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to update role")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "role updated"})
}

// Delete role
func (s *Server) deleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	var req ProjectRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}

	err := s.provider.DeleteRole(r.Context(), req.UserID, req.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrProjectNotFound):
			s.respondWithError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, users.ErrPermissionNotFound):
			s.respondWithError(w, http.StatusNotFound, "permission not found")
		case strings.Contains(err.Error(), "cannot delete owner"):
			s.respondWithError(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("DeleteRole error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to delete role")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "role deleted"})
}

// Responces:

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
