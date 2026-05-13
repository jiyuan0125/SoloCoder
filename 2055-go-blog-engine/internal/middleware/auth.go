package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"blog-engine/pkg/db"
	"blog-engine/pkg/utils"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessionID string

		cookie, err := r.Cookie("session_id")
		if err == nil {
			sessionID = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				sessionID = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if sessionID == "" {
			next.ServeHTTP(w, r)
			return
		}

		var userID int
		var expiresAt time.Time
		err = db.DB.QueryRow(
			`SELECT user_id, expires_at FROM sessions WHERE id = ?`,
			sessionID,
		).Scan(&userID, &expiresAt)

		if err == sql.ErrNoRows {
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if expiresAt.Before(time.Now()) {
			db.DB.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
			next.ServeHTTP(w, r)
			return
		}

		var user utils.UserInfo
		err = db.DB.QueryRow(
			`SELECT id, username, email, display_name, bio, role FROM users WHERE id = ?`,
			userID,
		).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName, &user.Bio, &user.Role)

		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), utils.UserKey, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(utils.UserKey).(*utils.UserInfo)
		if !ok || user == nil {
			utils.JSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(utils.UserKey).(*utils.UserInfo)
		if !ok || user == nil {
			utils.JSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
		if user.Role != "admin" {
			utils.JSONError(w, http.StatusForbidden, "Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
