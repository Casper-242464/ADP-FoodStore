package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"foodstore/internal/handlers"
	"foodstore/internal/repositories"
	"foodstore/internal/services"
)

func main() {
	// --- Database ---
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// --- Repos ---
	productRepo := repositories.NewProductRepository(db) // example
	orderRepo := repositories.NewOrderRepository(db)     // example
	contactRepo := repositories.NewContactRepository(db) // example

	// --- Services ---
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo)
	contactService := services.NewContactService(contactRepo)

	// --- Handlers (your existing layered handlers) ---
	ph := handlers.NewProductHandler(productService)
	oh := handlers.NewOrderHandler(orderService)
	ch := handlers.NewContactHandler(contactService)

	// ---------- API ----------
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/products", ph.ListProducts)   // JSON
	http.HandleFunc("/orders", oh.PlaceOrder)       // JSON
	http.HandleFunc("/contact", ch.HandleContact)   // GET HTML + POST form/JSON

	// ---------- UI ----------
	http.HandleFunc("/ui/products", handlers.ProductsPage)
	http.HandleFunc("/ui/orders", handlers.OrdersPage)
	http.HandleFunc("/ui/cart", handlers.CartPage)
	http.HandleFunc("/", handlers.HomePage) // keep last (catch-all for only "/")

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
