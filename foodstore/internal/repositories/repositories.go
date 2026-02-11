package repositories

import (
	"database/sql"
	"errors"
	"time"

	"foodstore/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

var ErrInsufficientStock = errors.New("insufficient stock")

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (pr *ProductRepository) GetAllProducts() ([]models.Product, error) {
	rows, err := pr.db.Query(
		"SELECT id, COALESCE(seller_id, 0), name, description, COALESCE(image_url, ''), price, stock, category, COALESCE(unit, 'piece'), created_at FROM products ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.ImageURL,
			&p.Price, &p.Stock, &p.Category, &p.Unit, &p.CreatedAt)
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

func (pr *ProductRepository) GetProductsBySellerID(sellerID int) ([]models.Product, error) {
	rows, err := pr.db.Query(
		"SELECT id, COALESCE(seller_id, 0), name, description, COALESCE(image_url, ''), price, stock, category, COALESCE(unit, 'piece'), created_at FROM products WHERE seller_id = $1 ORDER BY id DESC",
		sellerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.ImageURL,
			&p.Price, &p.Stock, &p.Category, &p.Unit, &p.CreatedAt)
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
		"INSERT INTO products (seller_id, name, description, image_url, price, stock, category, unit, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		p.SellerID, p.Name, p.Description, p.ImageURL, p.Price, p.Stock, p.Category, p.Unit, time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (pr *ProductRepository) UpdateProduct(p models.Product) (bool, error) {
	res, err := pr.db.Exec(
		"UPDATE products SET name=$1, description=$2, image_url=$3, price=$4, stock=$5, category=$6, unit=$7 WHERE id=$8 AND seller_id=$9",
		p.Name, p.Description, p.ImageURL, p.Price, p.Stock, p.Category, p.Unit, p.ID, p.SellerID,
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

func (pr *ProductRepository) UpdateProductAsAdmin(p models.Product) (bool, error) {
	res, err := pr.db.Exec(
		"UPDATE products SET name=$1, description=$2, image_url=$3, price=$4, stock=$5, category=$6, unit=$7 WHERE id=$8",
		p.Name, p.Description, p.ImageURL, p.Price, p.Stock, p.Category, p.Unit, p.ID,
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

func (pr *ProductRepository) DeleteProduct(id int, sellerID int) (bool, error) {
	res, err := pr.db.Exec("DELETE FROM products WHERE id=$1 AND seller_id=$2", id, sellerID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (pr *ProductRepository) DeleteProductAsAdmin(id int) (bool, error) {
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
		"SELECT id, COALESCE(seller_id, 0), name, description, COALESCE(image_url, ''), price, stock, category, COALESCE(unit, 'piece'), created_at FROM products WHERE id = $1", id)
	var p models.Product
	err := row.Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.ImageURL,
		&p.Price, &p.Stock, &p.Category, &p.Unit, &p.CreatedAt)
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

func (or *OrderRepository) CreateOrder(userID int, items []models.OrderItem, total float64, deliveryAddress, phoneNumber, comment string) (int, error) {
	tx, err := or.db.Begin()
	if err != nil {
		return 0, err
	}
	var orderID int
	err = tx.QueryRow(
		"INSERT INTO orders (user_id, total_price, status, delivery_address, phone_number, comment, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		userID, total, "pending", deliveryAddress, phoneNumber, comment, time.Now(),
	).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	for _, item := range items {
		res, err := tx.Exec(
			"UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1",
			item.Quantity, item.ProductID,
		)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		if affected == 0 {
			tx.Rollback()
			return 0, ErrInsufficientStock
		}

		_, err = tx.Exec(
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

func (or *OrderRepository) ListOrdersByUserID(userID int) ([]models.Order, error) {
	rows, err := or.db.Query(`
		SELECT
			o.id, o.user_id, o.total_price, o.status,
			COALESCE(o.delivery_address, ''), COALESCE(o.phone_number, ''), COALESCE(o.comment, ''), o.created_at,
			oi.id, oi.product_id, oi.quantity, oi.unit_price, oi.line_total,
			p.name
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN products p ON p.id = oi.product_id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC, o.id DESC, oi.id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[int]*models.Order)
	orderIDs := make([]int, 0)

	for rows.Next() {
		var o models.Order
		var itemID sql.NullInt64
		var productID sql.NullInt64
		var quantity sql.NullInt64
		var unitPrice sql.NullFloat64
		var lineTotal sql.NullFloat64
		var productName sql.NullString

		if err := rows.Scan(
			&o.ID, &o.UserID, &o.TotalPrice, &o.Status, &o.DeliveryAddress, &o.PhoneNumber, &o.Comment, &o.CreatedAt,
			&itemID, &productID, &quantity, &unitPrice, &lineTotal,
			&productName,
		); err != nil {
			return nil, err
		}

		existing, ok := orderMap[o.ID]
		if !ok {
			o.Items = []models.OrderItem{}
			orderMap[o.ID] = &o
			orderIDs = append(orderIDs, o.ID)
			existing = &o
		}

		if itemID.Valid {
			item := models.OrderItem{
				ID:          int(itemID.Int64),
				OrderID:     o.ID,
				ProductID:   int(productID.Int64),
				Quantity:    int(quantity.Int64),
				UnitPrice:   unitPrice.Float64,
				LineTotal:   lineTotal.Float64,
				ProductName: productName.String,
			}
			existing.Items = append(existing.Items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	orders := make([]models.Order, 0, len(orderIDs))
	for _, id := range orderIDs {
		if o := orderMap[id]; o != nil {
			orders = append(orders, *o)
		}
	}
	return orders, nil
}

func (or *OrderRepository) ListOrdersForSeller(sellerID int) ([]models.SellerOrder, error) {
	rows, err := or.db.Query(`
		SELECT
			o.id, o.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''), o.status,
			COALESCE(o.delivery_address, ''), COALESCE(o.phone_number, ''), COALESCE(o.comment, ''), o.created_at,
			oi.id, oi.product_id, oi.quantity, oi.unit_price, oi.line_total, p.name
		FROM orders o
		JOIN users u ON u.id = o.user_id
		JOIN order_items oi ON oi.order_id = o.id
		JOIN products p ON p.id = oi.product_id
		WHERE p.seller_id = $1
		ORDER BY o.created_at DESC, o.id DESC, oi.id ASC
	`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[int]*models.SellerOrder)
	orderIDs := make([]int, 0)

	for rows.Next() {
		var o models.SellerOrder
		var item models.OrderItem

		if err := rows.Scan(
			&o.ID, &o.UserID, &o.BuyerName, &o.BuyerEmail, &o.Status,
			&o.DeliveryAddress, &o.PhoneNumber, &o.Comment, &o.CreatedAt,
			&item.ID, &item.ProductID, &item.Quantity, &item.UnitPrice, &item.LineTotal, &item.ProductName,
		); err != nil {
			return nil, err
		}

		existing, ok := orderMap[o.ID]
		if !ok {
			o.Items = []models.OrderItem{}
			o.SellerTotal = 0
			orderMap[o.ID] = &o
			orderIDs = append(orderIDs, o.ID)
			existing = &o
		}

		item.OrderID = o.ID
		existing.Items = append(existing.Items, item)
		existing.SellerTotal += item.LineTotal
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	orders := make([]models.SellerOrder, 0, len(orderIDs))
	for _, id := range orderIDs {
		if o := orderMap[id]; o != nil {
			orders = append(orders, *o)
		}
	}
	return orders, nil
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

func (cr *ContactRepository) ListMessages() ([]models.ContactMessage, error) {
	rows, err := cr.db.Query(
		"SELECT id, COALESCE(user_id, 0), name, email, COALESCE(subject, ''), message, COALESCE(status, 'new'), COALESCE(created_at, NOW()) FROM contact_messages ORDER BY created_at DESC, id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]models.ContactMessage, 0)
	for rows.Next() {
		var msg models.ContactMessage
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Name, &msg.Email, &msg.Subject, &msg.Message, &msg.Status, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
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
		"SELECT id, name, email, COALESCE(password_hash, ''), COALESCE(role, 'buyer'), COALESCE(created_at, NOW()) FROM users WHERE email = $1", email,
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
		"SELECT id, name, email, COALESCE(password_hash, ''), COALESCE(role, 'buyer'), COALESCE(created_at, NOW()) FROM users WHERE id = $1", id,
	)
	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
