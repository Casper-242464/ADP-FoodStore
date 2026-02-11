package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"foodstore/internal/models"
	"foodstore/internal/services"
)

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
			UserID          int    `json:"user_id"`
			DeliveryAddress string `json:"delivery_address"`
			PhoneNumber     string `json:"phone_number"`
			Comment         string `json:"comment"`
			Items           []struct {
				ProductID int `json:"product_id"`
				Quantity  int `json:"quantity"`
			} `json:"items"`
		}
		body, err := io.ReadAll(r.Body)
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
		orderID, err := oh.service.PlaceOrder(reqBody.UserID, items, reqBody.DeliveryAddress, reqBody.PhoneNumber, reqBody.Comment)
		if err != nil {
			if errors.Is(err, services.ErrInvalidOrder) || errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrProductNotFound) || errors.Is(err, services.ErrInsufficientStock) {
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

func (oh *OrderHandler) SellerOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sellerID, err := getUserIDFromHeader(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing or invalid user id"})
		return
	}

	orders, err := oh.service.ListOrdersForSeller(sellerID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidOrder) || errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrSellerRequired) {
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
}
