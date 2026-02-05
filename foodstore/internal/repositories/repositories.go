package repositories

import (
	"database/sql"
	"time"

	"foodstore/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

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

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (or *OrderRepository) CreateOrder(userID int, items []models.OrderItem, total float64) (int, error) {
	tx, err := or.db.Begin()
	if err != nil {
		return 0, err
	}
	var orderID int
	err = tx.QueryRow(
		"INSERT INTO orders (user_id, total_price, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, total, "pending", time.Now(),
	).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
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
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return orderID, nil
}

type ContactRepository struct {
	db *sql.DB
}

func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (cr *ContactRepository) SaveMessage(msg models.ContactMessage) error {
	_, err := cr.db.Exec(
		"INSERT INTO contact_messages (user_id, name, email, subject, message, status, created_at) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7)",
		msg.UserID, msg.Name, msg.Email, msg.Subject, msg.Message, msg.Status, msg.CreatedAt,
	)
	return err
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) UserExists(id int) (bool, error) {
	var exists bool
	err := ur.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (ur *UserRepository) UserExistsByEmail(email string) (bool, error) {
	var exists bool
	err := ur.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (ur *UserRepository) CreateUser(user models.User) (int, error) {
	var id int
	err := ur.db.QueryRow(
		"INSERT INTO users (name, email, password_hash, role, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Name, user.Email, user.PasswordHash, user.Role, time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (ur *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	row := ur.db.QueryRow(
		"SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = $1", email,
	)
	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (ur *UserRepository) GetUserByID(id int) (*models.User, error) {
	row := ur.db.QueryRow(
		"SELECT id, name, email, password_hash, role, created_at FROM users WHERE id = $1", id,
	)
	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
