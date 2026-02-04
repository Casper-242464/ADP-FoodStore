package models

import "time"

// User represents a system user (customer or admin).
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"` // hashed password for authentication
	CreatedAt    time.Time `json:"created_at"`
}

// Product represents an item available in the store.
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`       // quantity in stock
	Category    string    `json:"category"`    // product category or type
	CreatedAt   time.Time `json:"created_at"`
}

// Order represents a purchase order made by a User.
type Order struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`      // e.g., "pending", "completed"
	CreatedAt  time.Time `json:"created_at"`
	// Note: Orders would typically include a list of OrderItems, but those are handled separately.
}

// OrderItem represents an individual item within an Order (linking a Product to an Order).
type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"order_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`  // price per unit at time of order
	LineTotal float64 `json:"line_total"`  // UnitPrice * Quantity
}

// ContactMessage represents a message sent by a user via the contact form.
type ContactMessage struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`   // optional: ID of logged-in user (if any)
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`    // e.g., "new", "handled"
	CreatedAt time.Time `json:"created_at"`
}
