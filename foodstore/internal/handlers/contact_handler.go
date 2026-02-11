package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"foodstore/internal/services"
)

type ContactHandler struct {
	service *services.ContactService
}

func NewContactHandler(cs *services.ContactService) *ContactHandler {
	return &ContactHandler{service: cs}
}

func (ch *ContactHandler) HandleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "frontend/pages/contacts.html")
		return
	}
	if r.Method == http.MethodPost {
		ct := r.Header.Get("Content-Type")
		var name, email, message string
		if strings.HasPrefix(ct, "application/json") {
			var reqBody struct {
				Name    string `json:"name"`
				Email   string `json:"email"`
				Message string `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
				return
			}
			name, email, message = reqBody.Name, reqBody.Email, reqBody.Message
		} else {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			name = r.FormValue("name")
			email = r.FormValue("email")
			message = r.FormValue("message")
		}
		if name == "" || email == "" || message == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "All fields are required"})
			return
		}
		userID := parseOptionalUserID(r)
		if err := ch.service.SendMessageFromUser(userID, name, email, message); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Thank you for contacting us",
		})
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (ch *ContactHandler) ListMessagesForAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adminID, err := getUserIDFromHeader(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing or invalid user id"})
		return
	}

	messages, err := ch.service.ListMessagesForAdmin(adminID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidOrder) || errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrAdminRequired) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func parseOptionalUserID(r *http.Request) int {
	userID, err := getUserIDFromHeader(r)
	if err != nil {
		return 0
	}
	return userID
}
