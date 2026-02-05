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
	cfg := config.GetConfig()
	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	productRepo := repositories.NewProductRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	userRepo := repositories.NewUserRepository(db)

	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo, userRepo)
	contactService := services.NewContactService(contactRepo)

	ph := handlers.NewProductHandler(productService)
	oh := handlers.NewOrderHandler(orderService)
	ch := handlers.NewContactHandler(contactService)

	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/products", ph.ListProducts)
	http.HandleFunc("/orders", oh.PlaceOrder)
	http.HandleFunc("/contact", ch.HandleContact)

	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("frontend/styles"))))

	http.HandleFunc("/ui/products", handlers.ProductsPage)
	http.HandleFunc("/ui/orders", handlers.OrdersPage)
	http.HandleFunc("/ui/cart", handlers.CartPage)
	http.HandleFunc("/", handlers.HomePage)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, nil))
}
