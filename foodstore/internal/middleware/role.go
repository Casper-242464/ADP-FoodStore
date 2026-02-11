package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"

	"foodstore/internal/services"
)

func RequireSeller(us *services.UserService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}
		if !requireSellerUser(w, r, us) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireSellerStrict(us *services.UserService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireSellerUser(w, r, us) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requireSellerUser(w http.ResponseWriter, r *http.Request, us *services.UserService) bool {
	userIDStr := r.Header.Get("X-User-Id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing or invalid user id"})
		return false
	}

	user, err := us.GetUserByID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return false
	}

	if user.Role != "seller" && user.Role != "administrator" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "seller or administrator role required"})
		return false
	}

	return true
}
