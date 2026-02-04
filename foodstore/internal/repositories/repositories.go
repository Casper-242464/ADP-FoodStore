package repositories

import (
	"database/sql"
	"time"

	"foodstore/internal/models"
)

// ProductRepository handles data operations for Product entities.
type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetAllProducts retrieves all products from the database.
func (pr *ProductRepository) GetAllProducts() ([]models.Product, error) {
	rows, err := pr.db.Query(
		"SELECT id, name, description, price, stock, category, created_at FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		// Scan each row into a Product struct
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price,
			&p.Stock, &p.Category, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

// CreateProduct inserts a new product and returns its ID.
func (pr *ProductRepository) CreateProduct(p models.Product) (int, error) {
	var id int
	err := pr.db.QueryRow(
		"INSERT INTO products (name, description, price, stock, category, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		p.Name, p.Description, p.Price, p.Stock, p.Category, time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateProduct updates an existing product by ID. Returns true if a row was updated.
func (pr *ProductRepository) UpdateProduct(p models.Product) (bool, error) {
	res, err := pr.db.Exec(
		"UPDATE products SET name=$1, description=$2, price=$3, stock=$4, category=$5 WHERE id=$6",
		p.Name, p.Description, p.Price, p.Stock, p.Category, p.ID,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// DeleteProduct deletes a product by ID. Returns true if a row was deleted.
func (pr *ProductRepository) DeleteProduct(id int) (bool, error) {
	res, err := pr.db.Exec("DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// GetProductByID retrieves a single product by its ID.
func (pr *ProductRepository) GetProductByID(id int) (*models.Product, error) {
	row := pr.db.QueryRow(
		"SELECT id, name, description, price, stock, category, created_at FROM products WHERE id = $1", id)
	var p models.Product
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Price,
		&p.Stock, &p.Category, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// OrderRepository handles data operations for orders (and related order items).
type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder inserts a new order and its items into the database. Returns the new Order ID.
func (or *OrderRepository) CreateOrder(userID int, items []models.OrderItem, total float64) (int, error) {
	tx, err := or.db.Begin()
	if err != nil {
		return 0, err
	}
	// Insert into orders table and get the generated Order ID.
	var orderID int
	err = tx.QueryRow(
		"INSERT INTO orders (user_id, total_price, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, total, "pending", time.Now(),
	).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	// Insert each item in the order_items table.
	for _, item := range items {
		_, err := tx.Exec(
			"INSERT INTO order_items (order_id, product_id, quantity, unit_price, line_total) VALUES ($1, $2, $3, $4, $5)",
			orderID, item.ProductID, item.Quantity, item.UnitPrice, item.LineTotal,
		)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	// Commit transaction if all inserts succeeded.
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return orderID, nil
}

// ContactRepository handles data operations for contact messages.
type ContactRepository struct {
	db *sql.DB
}

func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// SaveMessage saves a new contact message in the database.
func (cr *ContactRepository) SaveMessage(msg models.ContactMessage) error {
	_, err := cr.db.Exec(
		"INSERT INTO contact_messages (user_id, name, email, subject, message, status, created_at) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7)",
		msg.UserID, msg.Name, msg.Email, msg.Subject, msg.Message, msg.Status, msg.CreatedAt,
	)
	return err
}
