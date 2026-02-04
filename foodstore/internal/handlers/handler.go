package handlers

import (
	"encoding/json"
	// "fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"foodstore/internal/models"
	"foodstore/internal/services"
)

// HealthHandler responds to GET /health with a basic health status.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// ProductHandler holds a reference to ProductService.
type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(ps *services.ProductService) *ProductHandler {
	return &ProductHandler{service: ps}
}

// ListProducts handles GET /products.
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

// OrderHandler holds a reference to OrderService.
type OrderHandler struct {
	service *services.OrderService
}

func NewOrderHandler(os *services.OrderService) *OrderHandler {
	return &OrderHandler{service: os}
}

// PlaceOrder handles POST /orders.
func (oh *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Read and parse the JSON request body.
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
	// Convert request items to model.OrderItem slice for the service.
	items := make([]models.OrderItem, len(reqBody.Items))
	for i, item := range reqBody.Items {
		items[i] = models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	// Call the service to place the order.
	orderID, err := oh.service.PlaceOrder(reqBody.UserID, items)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	// Return the new Order ID as confirmation.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"order_id": orderID})
}

// ContactHandler holds a reference to ContactService.
type ContactHandler struct {
	service *services.ContactService
}

func NewContactHandler(cs *services.ContactService) *ContactHandler {
	return &ContactHandler{service: cs}
}

// HandleContact serves the contact page (GET) and processes submissions (POST).
func (ch *ContactHandler) HandleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Serve the static contact form page.
		http.ServeFile(w, r, "frontend/pages/contacts.html")
		return
	}
	if r.Method == http.MethodPost {
		// Determine content type to parse input accordingly.
		ct := r.Header.Get("Content-Type")
		var name, email, message string
		if strings.HasPrefix(ct, "application/json") {
			// If JSON, decode request body to get the fields.
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
			// Otherwise, assume form data (content-type: application/x-www-form-urlencoded).
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			name = r.FormValue("name")
			email = r.FormValue("email")
			message = r.FormValue("message")
		}
		// Basic validation: require all fields.
		if name == "" || email == "" || message == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "All fields are required"})
			return
		}
		// Use the service to handle the contact message.
		if err := ch.service.SendMessage(name, email, message); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		// Respond with a confirmation (JSON).
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Thank you for contacting us",
		})
		return
	}
	// If method is not GET or POST:
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
