package handlers

import (
	"encoding/json"
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
		if err := ch.service.SendMessage(name, email, message); err != nil {
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
