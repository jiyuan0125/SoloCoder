package middleware

import (
	"context"
	"net"
	"net/http"
)

type contextKey string

const (
	UserIDKey     contextKey = "user_id"
	IsAdminKey    contextKey = "is_admin"
	IPAddressKey  contextKey = "ip_address"
)

func getIPAddress(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "Unauthorized: missing X-User-ID header", http.StatusUnauthorized)
			return
		}

		isAdmin := r.Header.Get("X-Admin") == "true"
		ip := getIPAddress(r)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, IsAdminKey, isAdmin)
		ctx = context.WithValue(ctx, IPAddressKey, ip)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value(IsAdminKey).(bool)
		if !ok || !isAdmin {
			http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(UserIDKey).(string)
	return userID
}

func GetIsAdmin(r *http.Request) bool {
	isAdmin, _ := r.Context().Value(IsAdminKey).(bool)
	return isAdmin
}

func GetIPAddress(r *http.Request) string {
	ip, _ := r.Context().Value(IPAddressKey).(string)
	return ip
}
