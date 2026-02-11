package config

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	ServerAddress string
}

func GetConfig() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "123456789"),
		DBName:        getEnv("DB_NAME", "foodstore"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		ServerAddress: getEnv("SERVER_ADDR", ":8080"),
	}
}

func ConnectDB(cfg *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := ensureSchemaUpdates(db); err != nil {
		return nil, err
	}
	return db, nil
}

func ensureSchemaUpdates(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'buyer' CHECK (role IN ('buyer', 'seller', 'administrator')),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			seller_id INTEGER REFERENCES users(id),
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			image_url TEXT NOT NULL DEFAULT '',
			price NUMERIC(12,2) NOT NULL,
			stock INTEGER NOT NULL DEFAULT 0,
			category TEXT NOT NULL,
			unit TEXT NOT NULL DEFAULT 'piece',
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL REFERENCES users(id),
				total_price NUMERIC(12,2) NOT NULL,
				status TEXT NOT NULL,
				delivery_address TEXT NOT NULL DEFAULT '',
				phone_number TEXT NOT NULL DEFAULT '',
				comment TEXT NOT NULL DEFAULT '',
				created_at TIMESTAMP NOT NULL DEFAULT NOW()
			)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INTEGER NOT NULL REFERENCES orders(id),
			product_id INTEGER NOT NULL REFERENCES products(id),
			quantity INTEGER NOT NULL,
			unit_price NUMERIC(12,2) NOT NULL,
			line_total NUMERIC(12,2) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS contact_messages (
			id SERIAL PRIMARY KEY,
			user_id INTEGER,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			subject TEXT NOT NULL,
			message TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS password_hash TEXT`,
		`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS role TEXT`,
		`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS created_at TIMESTAMP`,
		`ALTER TABLE IF EXISTS products ADD COLUMN IF NOT EXISTS image_url TEXT`,
		`ALTER TABLE IF EXISTS products ADD COLUMN IF NOT EXISTS seller_id INTEGER`,
		`ALTER TABLE IF EXISTS products ADD COLUMN IF NOT EXISTS unit TEXT`,
		`ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS status TEXT`,
		`ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS delivery_address TEXT`,
		`ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS phone_number TEXT`,
		`ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS comment TEXT`,
		`ALTER TABLE IF EXISTS orders ADD COLUMN IF NOT EXISTS created_at TIMESTAMP`,
		`ALTER TABLE IF EXISTS order_items ADD COLUMN IF NOT EXISTS unit_price NUMERIC(12,2)`,
		`ALTER TABLE IF EXISTS order_items ADD COLUMN IF NOT EXISTS line_total NUMERIC(12,2)`,
		`ALTER TABLE IF EXISTS contact_messages ADD COLUMN IF NOT EXISTS subject TEXT`,
		`ALTER TABLE IF EXISTS contact_messages ADD COLUMN IF NOT EXISTS status TEXT`,
		`ALTER TABLE IF EXISTS contact_messages ADD COLUMN IF NOT EXISTS created_at TIMESTAMP`,
		`DO $$
			BEGIN
				IF EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_schema = 'public'
					  AND table_name = 'users'
					  AND column_name = 'password'
				) THEN
					EXECUTE 'UPDATE users SET password_hash = COALESCE(password_hash, password)';
				END IF;
			END
		$$`,
		`UPDATE users SET password_hash = '' WHERE password_hash IS NULL`,
		`UPDATE users SET role = 'buyer' WHERE role IS NULL OR role = '' OR role NOT IN ('buyer', 'seller', 'administrator')`,
		`UPDATE users SET created_at = NOW() WHERE created_at IS NULL`,
		`DO $$
				BEGIN
					IF NOT EXISTS (SELECT 1 FROM users WHERE role = 'administrator') THEN
						IF EXISTS (SELECT 1 FROM users WHERE email = 'admin@foodstore.local') THEN
							UPDATE users
							SET role = 'administrator',
								name = COALESCE(NULLIF(name, ''), 'Administrator'),
								password_hash = COALESCE(NULLIF(password_hash, ''), 'admin123')
							WHERE email = 'admin@foodstore.local';
						ELSE
							INSERT INTO users (name, email, password_hash, role, created_at)
							VALUES ('Administrator', 'admin@foodstore.local', 'admin123', 'administrator', NOW());
						END IF;
					END IF;
				END
			$$`,
		`UPDATE products SET image_url = '' WHERE image_url IS NULL`,
		`UPDATE products SET unit = 'piece' WHERE unit IS NULL OR unit = ''`,
		`UPDATE orders SET status = 'pending' WHERE status IS NULL OR status = ''`,
		`UPDATE orders SET delivery_address = '' WHERE delivery_address IS NULL`,
		`UPDATE orders SET phone_number = '' WHERE phone_number IS NULL`,
		`UPDATE orders SET comment = '' WHERE comment IS NULL`,
		`UPDATE orders SET created_at = NOW() WHERE created_at IS NULL`,
		`UPDATE order_items SET unit_price = 0 WHERE unit_price IS NULL`,
		`UPDATE order_items SET line_total = 0 WHERE line_total IS NULL`,
		`UPDATE contact_messages SET subject = '' WHERE subject IS NULL`,
		`UPDATE contact_messages SET status = 'new' WHERE status IS NULL OR status = ''`,
		`UPDATE contact_messages SET created_at = NOW() WHERE created_at IS NULL`,
		`ALTER TABLE IF EXISTS users ALTER COLUMN password_hash SET DEFAULT ''`,
		`ALTER TABLE IF EXISTS users ALTER COLUMN role SET DEFAULT 'buyer'`,
		`ALTER TABLE IF EXISTS users ALTER COLUMN created_at SET DEFAULT NOW()`,
		`ALTER TABLE IF EXISTS products ALTER COLUMN image_url SET DEFAULT ''`,
		`ALTER TABLE IF EXISTS products ALTER COLUMN unit SET DEFAULT 'piece'`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN status SET DEFAULT 'pending'`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN delivery_address SET DEFAULT ''`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN phone_number SET DEFAULT ''`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN comment SET DEFAULT ''`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN created_at SET DEFAULT NOW()`,
		`ALTER TABLE IF EXISTS contact_messages ALTER COLUMN subject SET DEFAULT ''`,
		`ALTER TABLE IF EXISTS contact_messages ALTER COLUMN status SET DEFAULT 'new'`,
		`ALTER TABLE IF EXISTS contact_messages ALTER COLUMN created_at SET DEFAULT NOW()`,
		`ALTER TABLE IF EXISTS users ALTER COLUMN password_hash SET NOT NULL`,
		`ALTER TABLE IF EXISTS users ALTER COLUMN role SET NOT NULL`,
		`ALTER TABLE IF EXISTS users ALTER COLUMN created_at SET NOT NULL`,
		`ALTER TABLE IF EXISTS products ALTER COLUMN image_url SET NOT NULL`,
		`ALTER TABLE IF EXISTS products ALTER COLUMN unit SET NOT NULL`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN status SET NOT NULL`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN delivery_address SET NOT NULL`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN phone_number SET NOT NULL`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN comment SET NOT NULL`,
		`ALTER TABLE IF EXISTS orders ALTER COLUMN created_at SET NOT NULL`,
		`ALTER TABLE IF EXISTS order_items ALTER COLUMN unit_price SET NOT NULL`,
		`ALTER TABLE IF EXISTS order_items ALTER COLUMN line_total SET NOT NULL`,
		`ALTER TABLE IF EXISTS contact_messages ALTER COLUMN subject SET NOT NULL`,
		`ALTER TABLE IF EXISTS contact_messages ALTER COLUMN status SET NOT NULL`,
		`ALTER TABLE IF EXISTS contact_messages ALTER COLUMN created_at SET NOT NULL`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("schema update failed: %w; statement: %s", err, stmt)
		}
	}
	return nil
}

func getEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
