package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"blog-engine/pkg/db"
	"blog-engine/pkg/utils"
)

type RegisterRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	DisplayName  string `json:"display_name"`
	Bio          string `json:"bio"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	Email       string `json:"email"`
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	Bio          string    `json:"bio"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		utils.JSONError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	var existingCount int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ? OR email = ?`, req.Username, req.Email).Scan(&existingCount)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if existingCount > 0 {
		utils.JSONError(w, http.StatusConflict, "Username or email already exists")
		return
	}

	passwordHash := utils.HashPassword(req.Password)
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	result, err := db.DB.Exec(
		`INSERT INTO users (username, email, password_hash, display_name, bio) VALUES (?, ?, ?, ?, ?)`,
		req.Username, req.Email, passwordHash, displayName, req.Bio,
	)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	userID, _ := result.LastInsertId()

	user := User{
		ID:          int(userID),
		Username:    req.Username,
		Email:       req.Email,
		DisplayName: displayName,
		Bio:         req.Bio,
		Role:        "author",
	}

	utils.JSONSuccess(w, http.StatusCreated, user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Identifier) == "" || strings.TrimSpace(req.Password) == "" {
		utils.JSONError(w, http.StatusBadRequest, "Identifier and password are required")
		return
	}

	var user User
	var passwordHash string
	err := db.DB.QueryRow(
		`SELECT id, username, email, display_name, bio, role, password_hash, created_at, updated_at 
		 FROM users WHERE username = ? OR email = ?`,
		req.Identifier, req.Identifier,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Bio, &user.Role, &passwordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if !utils.VerifyPassword(req.Password, passwordHash) {
		utils.JSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	sessionID := utils.GenerateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = db.DB.Exec(
		`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		sessionID, user.ID, expiresAt,
	)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Expires:  expiresAt,
		SameSite: http.SameSiteLaxMode,
	})

	utils.JSONSuccess(w, http.StatusOK, map[string]interface{}{
		"user":    user,
		"session": sessionID,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		db.DB.Exec(`DELETE FROM sessions WHERE id = ?`, cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	utils.JSONSuccess(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := r.Context().Value(utils.UserKey).(*utils.UserInfo)
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, userInfo)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := r.Context().Value(utils.UserKey).(*utils.UserInfo)
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email != "" && req.Email != userInfo.Email {
		var existingCount int
		err := db.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ? AND id != ?`, req.Email, userInfo.ID).Scan(&existingCount)
		if err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "Database error")
			return
		}
		if existingCount > 0 {
			utils.JSONError(w, http.StatusConflict, "Email already exists")
			return
		}
	}

	updates := []string{}
	args := []interface{}{}

	if req.DisplayName != "" {
		updates = append(updates, "display_name = ?")
		args = append(args, req.DisplayName)
	}
	if req.Bio != "" {
		updates = append(updates, "bio = ?")
		args = append(args, req.Bio)
	}
	if req.Email != "" {
		updates = append(updates, "email = ?")
		args = append(args, req.Email)
	}

	if len(updates) == 0 {
		utils.JSONError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userInfo.ID)

	query := `UPDATE users SET ` + strings.Join(updates, ", ") + ` WHERE id = ?`
	_, err := db.DB.Exec(query, args...)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, map[string]string{"message": "Profile updated successfully"})
}

func GetUserByID(id int) (*utils.UserInfo, error) {
	var user utils.UserInfo
	err := db.DB.QueryRow(
		`SELECT id, username, email, display_name, bio, role FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Bio, &user.Role)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func InitAdmin() error {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	passwordHash := utils.HashPassword("admin123")
	_, err = db.DB.Exec(
		`INSERT INTO users (username, email, password_hash, display_name, bio, role) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"admin", "admin@example.com", passwordHash, "Admin", "System Administrator", "admin",
	)
	return err
}

func UserFromContext(ctx context.Context) *utils.UserInfo {
	user, ok := ctx.Value(utils.UserKey).(*utils.UserInfo)
	if !ok {
		return nil
	}
	return user
}
