# Demo Steps (Milestone 2)

## 1) Start database and load schema
```
psql -U postgres -d foodstore -f "/Users/bekasyljaksylyk/работа/advanced programming1/assignment3/foodstore/schema.sql"
```

## 1.1) Seed a user (required for orders)
```
psql -U postgres -d foodstore -c "INSERT INTO users (name,email,password_hash) VALUES ('Demo User','demo@example.com','x') RETURNING id;"
```

## 2) Run backend
```
DB_USER=postgres DB_PASSWORD=1109 DB_NAME=foodstore DB_SSLMODE=disable go run .
```

## 3) Endpoints to show (JSON)
```
curl http://localhost:8080/health
```

```
curl http://localhost:8080/products
```

```
curl -X POST http://localhost:8080/products \
  -H "X-User-Id: 1" \
  -F "name=Apple" \
  -F "description=Fresh" \
  -F "unit=kg" \
  -F "price=1.50" \
  -F "stock=10" \
  -F "category=Fruit" \
  -F "image=@/absolute/path/to/apple.jpg"
```

```
curl -X PUT http://localhost:8080/products \
  -H "X-User-Id: 1" \
  -F "id=1" \
  -F "name=Apple" \
  -F "description=Fresh" \
  -F "unit=kg" \
  -F "price=1.75" \
  -F "stock=12" \
  -F "category=Fruit"
```

```
curl -X DELETE "http://localhost:8080/products?id=1" \
  -H "X-User-Id: 1"
```

```
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"items":[{"product_id":2,"quantity":1}]}'
```

```
curl -X POST http://localhost:8080/contact \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","message":"Hello"}'
```

## 4) Frontend demo (optional)
- http://localhost:8080/ui/products
- http://localhost:8080/ui/seller/products
- http://localhost:8080/ui/orders
