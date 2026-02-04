package services

import (
	"log"
	"time"

	"foodstore/internal/models"
	"foodstore/internal/repositories"
)

// ProductService provides business logic for products.
type ProductService struct {
	productRepo *repositories.ProductRepository
}

func NewProductService(pr *repositories.ProductRepository) *ProductService {
	return &ProductService{productRepo: pr}
}

// ListProducts returns all products (business logic could filter or enhance data).
func (ps *ProductService) ListProducts() ([]models.Product, error) {
	products, err := ps.productRepo.GetAllProducts()
	if err != nil {
		return nil, err
	}
	// (Additional logic like sorting or filtering could be applied here if needed)
	return products, nil
}

// CreateProduct validates and creates a new product.
func (ps *ProductService) CreateProduct(p models.Product) (int, error) {
	return ps.productRepo.CreateProduct(p)
}

// UpdateProduct validates and updates an existing product.
func (ps *ProductService) UpdateProduct(p models.Product) (bool, error) {
	return ps.productRepo.UpdateProduct(p)
}

// DeleteProduct deletes a product by ID.
func (ps *ProductService) DeleteProduct(id int) (bool, error) {
	return ps.productRepo.DeleteProduct(id)
}

// OrderService provides business logic for orders.
type OrderService struct {
	orderRepo   *repositories.OrderRepository
	productRepo *repositories.ProductRepository // used to fetch product details for calculations
}

func NewOrderService(or *repositories.OrderRepository, pr *repositories.ProductRepository) *OrderService {
	return &OrderService{orderRepo: or, productRepo: pr}
}

// PlaceOrder processes a new order request: calculates totals and saves the order.
func (os *OrderService) PlaceOrder(userID int, items []models.OrderItem) (int, error) {
	// Calculate the total price and populate item prices using product data.
	var total float64
	for i := range items {
		// Look up each product's current price
		product, err := os.productRepo.GetProductByID(items[i].ProductID)
		if err != nil {
			return 0, err // if product not found or DB error
		}
		items[i].UnitPrice = product.Price
		items[i].LineTotal = product.Price * float64(items[i].Quantity)
		total += items[i].LineTotal
		// (Potential business rule: check if product.Stock is enough for Quantity, etc.)
	}
	// Call repository to create the order in the database (within a transaction).
	orderID, err := os.orderRepo.CreateOrder(userID, items, total)
	if err != nil {
		return 0, err
	}
	// (In a full implementation, we might adjust product stock levels here as well.)
	return orderID, nil
}

// ContactService provides business logic for contact form messages.
type ContactService struct {
	contactRepo *repositories.ContactRepository
}

func NewContactService(cr *repositories.ContactRepository) *ContactService {
	return &ContactService{contactRepo: cr}
}

// SendMessage processes a new contact message: saves it and triggers background notification.
func (cs *ContactService) SendMessage(name, email, message string) error {
	// Create a ContactMessage model instance.
	msg := models.ContactMessage{
		UserID:    0, // assume not logged in (or set user ID if available)
		Name:      name,
		Email:     email,
		Subject:   "", // subject could be set if provided
		Message:   message,
		Status:    "new",
		CreatedAt: time.Now(),
	}
	// Save the message via the repository.
	if err := cs.contactRepo.SaveMessage(msg); err != nil {
		return err
	}
	// Launch a background goroutine to simulate sending a notification email.
	go func(m models.ContactMessage) {
		// Simulate email sending delay
		time.Sleep(2 * time.Second)
		log.Printf("Background: Sent email notification for new contact message from %s", m.Email)
	}(msg)
	return nil
}
