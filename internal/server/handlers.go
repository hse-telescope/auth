package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

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

//////////////////
//				//
//	Auth base	//
//				//
//////////////////

// Login user
func (s *Server) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeaders(w)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Default().Println("\n---LOGIN---\n[REQUEST]: ", req)

	tokens, err := s.provider.LoginUser(context.Background(), req.LoginData, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			http.Error(w, "User not found", http.StatusNotFound)
		case errors.Is(err, users.ErrIncorrectPassword):
			http.Error(w, "Incorrect password", http.StatusUnauthorized)
		default:
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

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	log.Default().Println("\n---REGISTER---\n[REQUEST]: ", req)

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tokens, err := s.provider.RegisterUser(context.Background(), req.Username, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUsernameExists):
			http.Error(w, "Username already exists", http.StatusConflict)
		case errors.Is(err, users.ErrEmailExists):
			http.Error(w, "Email already exists", http.StatusConflict)
		default:
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
	log.Default().Println("\n---LOGOUT---\n[REQUEST]: ", req)

	response := map[string]string{
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

//////////////////////
//					//
//	Role managment	//
//					//
//////////////////////

// Get user projects handler
func (s *Server) getUserProjectsHandler(w http.ResponseWriter, r *http.Request) {
	var req GetUserProjectsRequest
	var err error

	req.UserID, err = strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}

	log.Default().Println("\n---GET PROJECTS---\n[REQUEST]: ", req)

	projectIDs, err := s.provider.GetUserProjects(context.Background(), req.UserID)
	if err != nil {
		s.respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"project_ids": projectIDs,
	})
}

// Create project
func (s *Server) createProjectHandler(w http.ResponseWriter, r *http.Request) {

	var req CreateProjectPermissionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	log.Default().Println("\n---CREATE PROJECT---\n[REQUEST]: ", req)

	err := s.provider.CreateProject(r.Context(), req.UserID, req.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrProjectExists):
			s.respondWithError(w, http.StatusConflict, "project already exists")
		case errors.Is(err, users.ErrPermissionExists):
			s.respondWithError(w, http.StatusConflict, "project permission already exists")
		default:
			s.respondWithError(w, http.StatusInternalServerError, "failed to create project: "+err.Error())
		}
		return
	}

	s.respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"project_id": req.ProjectID,
		"status":     "success",
	})
}

// Get users role in project
type Role int

const (
	Viewer Role = iota
	Editor
	Owner
)

func parseRole(roleStr string) (Role, bool) {
	switch roleStr {
	case "viewer":
		return Viewer, true
	case "editor":
		return Editor, true
	case "owner":
		return Owner, true
	default:
		return Viewer, false
	}
}

func hasPermission(currRoleStr, reqRoleStr string) bool {
	currRole, ok1 := parseRole(currRoleStr)
	reqRole, ok2 := parseRole(reqRoleStr)
	if !ok1 || !ok2 {
		return false
	}
	return currRole >= reqRole
}

func (s *Server) getUserProjectRoleHandler(w http.ResponseWriter, r *http.Request) {
	var req GetRoleRequest
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

	log.Default().Println("\n---GET ROLE---\n[REQUEST]: ", req)

	currRole, err := s.provider.GetRole(r.Context(), req.UserID, req.ProjectID)
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

	isRoleEnough := hasPermission(currRole, req.Role)

	s.respondWithJSON(w, http.StatusOK, map[string]bool{"isRoleEnough": isRoleEnough})
}

// Assign role
func (s *Server) assignRoleHandler(w http.ResponseWriter, r *http.Request) {

	var req AssignRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	log.Default().Println("\n---ASSIGN ROLE---\n[REQUEST]: ", req)

	err := s.provider.AssignRole(r.Context(), req.UserID, req.Username, req.ProjectID, req.Role)
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
		case errors.Is(err, users.ErrAssignerIsNotOwner):
			s.respondWithError(w, http.StatusConflict, "assigner is not owner")
		case errors.Is(err, users.ErrAssignerRoleNotFound):
			s.respondWithError(w, http.StatusNotFound, "assigner role not found")
		case errors.Is(err, users.ErrAssignableNotFound):
			s.respondWithError(w, http.StatusNotFound, "assignable not found")
		case errors.Is(err, users.ErrOwnerRoleChanging):
			s.respondWithError(w, http.StatusForbidden, "cannot assign role to owner")
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
	var req UpdateRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	log.Default().Println("\n---UPDATE ROLE---\n[REQUEST]: ", req)

	err := s.provider.UpdateRole(r.Context(), req.UserID, req.Username, req.ProjectID, req.Role)
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
		case errors.Is(err, users.ErrPermissionExists):
			s.respondWithError(w, http.StatusForbidden, "assigner is not owner")
		case errors.Is(err, users.ErrAssignerRoleNotFound):
			s.respondWithError(w, http.StatusNotFound, "assigner role not found")
		case errors.Is(err, users.ErrAssignableNotFound):
			s.respondWithError(w, http.StatusNotFound, "assignable not found")
		case errors.Is(err, users.ErrAssignableRoleNotFound):
			s.respondWithError(w, http.StatusNotFound, "assignable role not found")
		case errors.Is(err, users.ErrOwnerRoleChanging):
			s.respondWithError(w, http.StatusForbidden, "cannot update owner role")
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
	var req DeleteRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	log.Default().Println("\n---DELETE ROLE---\n[REQUEST]: ", req)

	err := s.provider.DeleteRole(r.Context(), req.UserID, req.Username, req.ProjectID)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrProjectNotFound):
			s.respondWithError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, users.ErrPermissionNotFound):
			s.respondWithError(w, http.StatusNotFound, "permission not found")
		case errors.Is(err, users.ErrPermissionExists):
			s.respondWithError(w, http.StatusConflict, "role already assigned")
		case errors.Is(err, users.ErrAssignerIsNotOwner):
			s.respondWithError(w, http.StatusConflict, "assigner is not owner")
		case errors.Is(err, users.ErrAssignerRoleNotFound):
			s.respondWithError(w, http.StatusNotFound, "assigner role not found")
		case errors.Is(err, users.ErrAssignableNotFound):
			s.respondWithError(w, http.StatusNotFound, "assignable not found")
		case errors.Is(err, users.ErrAssignableRoleNotFound):
			s.respondWithError(w, http.StatusNotFound, "assignable role not found")
		case errors.Is(err, users.ErrOwnerRoleChanging):
			s.respondWithError(w, http.StatusForbidden, "cannot delete owner role")
		default:
			log.Printf("DeleteRole error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to delete role")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "role deleted"})
}

///////////////////////
//					 //
// Change login data //
//					 //
///////////////////////

func (s *Server) changeUsernameHandler(w http.ResponseWriter, r *http.Request) {
	var req ChangeUsernameRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	log.Default().Println("\n---CHANGE USERNAME---\n[REQUEST]: ", req)
	log.Default().Printf("%s\n%s\n%s\n%s\n", req.Email, req.NewUsername, req.OldUsername, req.Password)

	err := s.provider.ChangeUsername(r.Context(), req.OldUsername, req.NewUsername, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrIncorrectPassword):
			s.respondWithError(w, http.StatusUnauthorized, "incorrect password")
		case errors.Is(err, users.ErrUsernameExists):
			s.respondWithError(w, http.StatusConflict, "username already exists")
		default:
			log.Printf("Change username error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to change username")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "username changed"})
}

func (s *Server) changeEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req ChangeEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	log.Default().Println("\n---CHANGE EMAIL---\n[REQUEST]: ", req)

	err := s.provider.ChangeEmail(r.Context(), req.Username, req.OldEmail, req.NewEmail, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrIncorrectPassword):
			s.respondWithError(w, http.StatusUnauthorized, "incorrect password")
		case errors.Is(err, users.ErrEmailExists):
			s.respondWithError(w, http.StatusConflict, "email already exists")
		default:
			log.Printf("Change email error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to change email")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "email changed"})
}

func (s *Server) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	log.Default().Println("\n---CHANGE PASSWORD---\n[REQUEST]: ", req)

	err := s.provider.ChangePassword(r.Context(), req.Username, req.Email, req.OldPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, users.ErrIncorrectPassword):
			s.respondWithError(w, http.StatusUnauthorized, "incorrect password")
		default:
			log.Printf("Change password error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to change password")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}

func (s *Server) forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}
	log.Default().Println("\n---FORGOT PASSWORD---\n[REQUEST]: ", req)

	err := s.provider.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			s.respondWithError(w, http.StatusNotFound, "user not found")
		default:
			log.Printf("Reset password error: %v", err)
			s.respondWithError(w, http.StatusInternalServerError, "failed to reset password")
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "password has been sent"})
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
