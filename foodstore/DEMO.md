# Demo Steps (Milestone 2)

## 1) Start database and load schema
```
psql -U postgres -d foodstore -f "/Users/bekasyljaksylyk/работа/advanced programming1/assignment3/foodstore/schema.sql"
```

## 2) Run backend
```
DB_USER=postgres DB_PASSWORD=123456789 DB_NAME=foodstore DB_SSLMODE=disable go run .
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
  -H "Content-Type: application/json" \
  -d '{"name":"Apple","description":"Fresh","price":1.50,"stock":10,"category":"Fruit"}'
```

```
curl -X PUT http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"id":1,"name":"Apple","description":"Fresh","price":1.75,"stock":12,"category":"Fruit"}'
```

```
curl -X DELETE "http://localhost:8080/products?id=1"
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
- http://localhost:8080/ui/orders
