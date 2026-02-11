package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Product struct {
	ID          int       `json:"id"`
	SellerID    int       `json:"seller_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	Unit        string    `json:"unit"`
	CreatedAt   time.Time `json:"created_at"`
}

type Order struct {
	ID              int         `json:"id"`
	UserID          int         `json:"user_id"`
	TotalPrice      float64     `json:"total_price"`
	Status          string      `json:"status"`
	DeliveryAddress string      `json:"delivery_address"`
	PhoneNumber     string      `json:"phone_number"`
	Comment         string      `json:"comment"`
	CreatedAt       time.Time   `json:"created_at"`
	Items           []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID          int     `json:"id"`
	OrderID     int     `json:"order_id"`
	ProductID   int     `json:"product_id"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	LineTotal   float64 `json:"line_total"`
	ProductName string  `json:"name"`
}

type ContactMessage struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type SellerOrder struct {
	ID              int         `json:"id"`
	UserID          int         `json:"user_id"`
	BuyerName       string      `json:"buyer_name"`
	BuyerEmail      string      `json:"buyer_email"`
	Status          string      `json:"status"`
	DeliveryAddress string      `json:"delivery_address"`
	PhoneNumber     string      `json:"phone_number"`
	Comment         string      `json:"comment"`
	SellerTotal     float64     `json:"seller_total"`
	CreatedAt       time.Time   `json:"created_at"`
	Items           []OrderItem `json:"items,omitempty"`
}
