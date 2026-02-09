package handlers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"foodstore/internal/models"
	"foodstore/internal/services"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(ps *services.ProductService) *ProductHandler {
	return &ProductHandler{service: ps}
}

func (ph *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		products, err := ph.service.ListProducts()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	case http.MethodPost:
		var reqBody struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Stock       int     `json:"stock"`
			Category    string  `json:"category"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}
		if reqBody.Name == "" || reqBody.Description == "" || reqBody.Category == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "name, description, category are required"})
			return
		}
		id, err := ph.service.CreateProduct(models.Product{
			Name:        reqBody.Name,
			Description: reqBody.Description,
			Price:       reqBody.Price,
			Stock:       reqBody.Stock,
			Category:    reqBody.Category,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	case http.MethodPut:
		var reqBody struct {
			ID          int     `json:"id"`
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Stock       int     `json:"stock"`
			Category    string  `json:"category"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}
		if reqBody.ID <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "id is required"})
			return
		}
		updated, err := ph.service.UpdateProduct(models.Product{
			ID:          reqBody.ID,
			Name:        reqBody.Name,
			Description: reqBody.Description,
			Price:       reqBody.Price,
			Stock:       reqBody.Stock,
			Category:    reqBody.Category,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if !updated {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "product not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}
		deleted, err := ph.service.DeleteProduct(id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if !deleted {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "product not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler(os *services.OrderService) *OrderHandler {
	return &OrderHandler{service: os}
}

func (oh *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "user_id is required"})
			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil || userID <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid user_id"})
			return
		}
		orders, err := oh.service.ListOrdersByUserID(userID)
		if err != nil {
			if errors.Is(err, services.ErrInvalidOrder) || errors.Is(err, services.ErrUserNotFound) {
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
		json.NewEncoder(w).Encode(orders)
		return
	case http.MethodPost:
		var reqBody struct {
			UserID int `json:"user_id"`
			Items  []struct {
				ProductID int `json:"product_id"`
				Quantity  int `json:"quantity"`
			} `json:"items"`
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &reqBody); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON format"})
			return
		}
		items := make([]models.OrderItem, len(reqBody.Items))
		for i, item := range reqBody.Items {
			items[i] = models.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
		}
		orderID, err := oh.service.PlaceOrder(reqBody.UserID, items)
		if err != nil {
			if errors.Is(err, services.ErrInvalidOrder) || errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrProductNotFound) {
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
		json.NewEncoder(w).Encode(map[string]interface{}{"order_id": orderID})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

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

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(us *services.UserService) *UserHandler {
	return &UserHandler{service: us}
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var reqBody struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	id, err := uh.service.Register(reqBody.Name, reqBody.Email, reqBody.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"user_id": id,
		"message": "Registration successful",
	})
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	user, err := uh.service.Login(reqBody.Email, reqBody.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (uh *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user id is required"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user id"})
		return
	}

	user, err := uh.service.GetUserByID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"user": map[string]interface{}{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
	})
}
