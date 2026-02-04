package main

import (
	"log"
	"net/http"

	"foodstore/config"
	"foodstore/internal/handlers"
	"foodstore/internal/repositories"
	"foodstore/internal/services"
)

func main() {
	// --- Database ---
	cfg := config.GetConfig()
	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// --- Repos ---
	productRepo := repositories.NewProductRepository(db) // example
	orderRepo := repositories.NewOrderRepository(db)     // example
	contactRepo := repositories.NewContactRepository(db) // example
	userRepo := repositories.NewUserRepository(db)       // example

	// --- Services ---
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo, userRepo)
	contactService := services.NewContactService(contactRepo)

	// --- Handlers (your existing layered handlers) ---
	ph := handlers.NewProductHandler(productService)
	oh := handlers.NewOrderHandler(orderService)
	ch := handlers.NewContactHandler(contactService)

	// ---------- API ----------
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/products", ph.ListProducts) // JSON
	http.HandleFunc("/orders", oh.PlaceOrder)     // JSON
	http.HandleFunc("/contact", ch.HandleContact) // GET HTML + POST form/JSON

	// ---------- UI ----------
	http.HandleFunc("/ui/products", handlers.ProductsPage)
	http.HandleFunc("/ui/orders", handlers.OrdersPage)
	http.HandleFunc("/ui/cart", handlers.CartPage)
	http.HandleFunc("/", handlers.HomePage) // keep last (catch-all for only "/")

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, nil))
}
