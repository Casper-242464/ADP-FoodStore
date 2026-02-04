# Food Store Backend + UI

A clean Go monolith for the Assignment 3 / Milestone 2 check.
It ships a working HTTP server, Postgres persistence, core domain models,
and a minimal UI to drive the JSON API from a browser.

## Highlights
- Net/http server with JSON endpoints and HTML pages.
- Core models: users, products, orders, order_items, contact_messages.
- Product CRUD (GET/POST/PUT/DELETE) and order creation flow.
- Postgres persistence with schema and foreign keys.
- Background goroutine for async contact notification.

## Architecture (Simple Monolith)
```
main.go
  -> handlers (HTTP, JSON, pages)
  -> services (business logic + validation)
  -> repositories (DB access)
  -> models (domain types)
```

## Requirements
- Go 1.22+
- Postgres 13+

## Quick Start
1) Load DB schema
```
psql -U postgres -d foodstore -f "/Users/bekasyljaksylyk/работа/advanced programming1/assignment3/foodstore/schema.sql"
```

2) Seed a user (required for orders)
```
psql -U postgres -d foodstore -c "INSERT INTO users (name,email,password_hash) VALUES ('Demo User','demo@example.com','x') RETURNING id;"
```

3) Run server
```
DB_USER=postgres DB_PASSWORD=123456789 DB_NAME=foodstore DB_SSLMODE=disable go run .
```

Server runs at:
```
http://localhost:8080
```

## Environment Variables
- DB_HOST (default: localhost)
- DB_PORT (default: 5432)
- DB_USER (default: postgres)
- DB_PASSWORD (default: 123456789)
- DB_NAME (default: foodstore)
- DB_SSLMODE (default: disable)
- SERVER_ADDR (default: :8080)

## Core API (JSON)
Health:
```
GET /health
```

Products CRUD:
```
GET    /products
POST   /products
PUT    /products
DELETE /products?id=1
```

Orders:
```
POST /orders
```

Contact:
```
POST /contact
```

## Sample Requests
Create product:
```
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Apple","description":"Fresh","price":1.5,"stock":10,"category":"Fruit"}'
```

Update product:
```
curl -X PUT http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"id":1,"name":"Apple","description":"Fresh","price":1.75,"stock":12,"category":"Fruit"}'
```

Delete product:
```
curl -X DELETE "http://localhost:8080/products?id=1"
```

Place order:
```
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"items":[{"product_id":2,"quantity":1},{"product_id":3,"quantity":2}]}'
```

Contact:
```
curl -X POST http://localhost:8080/contact \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","message":"Hello"}'
```

## UI Pages (Optional)
- /ui/products (create + list products)
- /ui/orders (place order)
- /ui/cart

Orders page input format:
- User ID: existing user id (e.g. 1)
- Items: "1:1, 2, 3:2" (if qty is missing, it defaults to 1)

## Milestone 2 Checklist Mapping
- Backend app: net/http server in `main.go`
- >=3 endpoints: /health, /products, /orders, /contact
- JSON input/output: orders/products/contact
- Data model: models + schema.sql
- CRUD: full CRUD for products
- Persistence: Postgres repositories
- Concurrency: goroutine in ContactService

## Troubleshooting
- "relation does not exist": run schema.sql.
- "SSL is not enabled": use DB_SSLMODE=disable.
- "user not found": seed a user and use its id.

## Files
- `schema.sql` - database schema
- `DEMO.md` - demo steps for presentation
- `internal/models` - core domain types
- `internal/repositories` - DB access
- `internal/services` - business logic
- `internal/handlers` - HTTP handlers
