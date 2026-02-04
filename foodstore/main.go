package main

import (
	"fmt"
	"log"
	"net/http"

	"foodstore/config"
	"foodstore/internal/handlers"
	"foodstore/internal/repositories"
	"foodstore/internal/services"
)

func main() {
	// Load configuration (from env variables or defaults)
	cfg := config.GetConfig()

	// Connect to the PostgreSQL database
	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories with the DB connection
	productRepo := repositories.NewProductRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	contactRepo := repositories.NewContactRepository(db)

	// Initialize services with repositories
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo)
	contactService := services.NewContactService(contactRepo)

	// Initialize handlers with services
	productHandler := handlers.NewProductHandler(productService)
	orderHandler := handlers.NewOrderHandler(orderService)
	contactHandler := handlers.NewContactHandler(contactService)

	// Set up HTTP routes and handlers
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/products", productHandler.ListProducts)   // GET
	http.HandleFunc("/orders", orderHandler.PlaceOrder)        // POST
	http.HandleFunc("/contact", contactHandler.HandleContact)  // GET and POST

	// Serve the homepage at "/" (if any other static file requested under root, return 404)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "frontend/pages/home.html")
	})

	// Start the HTTP server
	addr := cfg.ServerAddress  // typically ":8080"
	fmt.Printf("Starting server on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
