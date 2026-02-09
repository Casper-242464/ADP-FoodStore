package services

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"foodstore/internal/models"
	"foodstore/internal/repositories"
)

type ProductService struct {
	productRepo *repositories.ProductRepository
}

func NewProductService(pr *repositories.ProductRepository) *ProductService {
	return &ProductService{productRepo: pr}
}

func (ps *ProductService) ListProducts() ([]models.Product, error) {
	products, err := ps.productRepo.GetAllProducts()
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (ps *ProductService) CreateProduct(p models.Product) (int, error) {
	return ps.productRepo.CreateProduct(p)
}

func (ps *ProductService) UpdateProduct(p models.Product) (bool, error) {
	return ps.productRepo.UpdateProduct(p)
}

func (ps *ProductService) DeleteProduct(id int) (bool, error) {
	return ps.productRepo.DeleteProduct(id)
}

type OrderService struct {
	orderRepo   *repositories.OrderRepository
	productRepo *repositories.ProductRepository
	userRepo    *repositories.UserRepository
}

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidOrder    = errors.New("invalid order")
)

func NewOrderService(or *repositories.OrderRepository, pr *repositories.ProductRepository, ur *repositories.UserRepository) *OrderService {
	return &OrderService{orderRepo: or, productRepo: pr, userRepo: ur}
}

func (os *OrderService) PlaceOrder(userID int, items []models.OrderItem) (int, error) {
	if userID <= 0 || len(items) == 0 {
		return 0, ErrInvalidOrder
	}
	exists, err := os.userRepo.UserExists(userID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, ErrUserNotFound
	}
	var total float64
	for i := range items {
		if items[i].ProductID <= 0 || items[i].Quantity <= 0 {
			return 0, ErrInvalidOrder
		}
		product, err := os.productRepo.GetProductByID(items[i].ProductID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, ErrProductNotFound
			}
			return 0, err
		}
		items[i].UnitPrice = product.Price
		items[i].LineTotal = product.Price * float64(items[i].Quantity)
		total += items[i].LineTotal
	}
	orderID, err := os.orderRepo.CreateOrder(userID, items, total)
	if err != nil {
		return 0, err
	}
	return orderID, nil
}

func (os *OrderService) ListOrdersByUserID(userID int) ([]models.Order, error) {
	if userID <= 0 {
		return nil, ErrInvalidOrder
	}
	exists, err := os.userRepo.UserExists(userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}
	return os.orderRepo.ListOrdersByUserID(userID)
}

type ContactService struct {
	contactRepo *repositories.ContactRepository
}

func NewContactService(cr *repositories.ContactRepository) *ContactService {
	return &ContactService{contactRepo: cr}
}

func (cs *ContactService) SendMessage(name, email, message string) error {
	msg := models.ContactMessage{
		UserID:    0,
		Name:      name,
		Email:     email,
		Subject:   "",
		Message:   message,
		Status:    "new",
		CreatedAt: time.Now(),
	}
	if err := cs.contactRepo.SaveMessage(msg); err != nil {
		return err
	}
	go func(m models.ContactMessage) {
		time.Sleep(2 * time.Second)
		log.Printf("Background: Sent email notification for new contact message from %s", m.Email)
	}(msg)
	return nil
}

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(ur *repositories.UserRepository) *UserService {
	return &UserService{userRepo: ur}
}

var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRole        = errors.New("invalid role")
)

func (us *UserService) Register(name, email, password, role string) (int, string, error) {
	if name == "" || email == "" || password == "" {
		return 0, "", errors.New("name, email, and password are required")
	}

	exists, err := us.userRepo.UserExistsByEmail(email)
	if err != nil {
		return 0, "", err
	}
	if exists {
		return 0, "", ErrUserAlreadyExists
	}

	if role == "" {
		role = "buyer"
	}
	if role != "buyer" && role != "seller" {
		return 0, "", ErrInvalidRole
	}

	user := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: password, 
		Role:         role,
	}
	id, err := us.userRepo.CreateUser(user)
	if err != nil {
		return 0, "", err
	}
	return id, role, nil
}

func (us *UserService) Login(email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash != password {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (us *UserService) GetUserByID(id int) (*models.User, error) {
	return us.userRepo.GetUserByID(id)
}
