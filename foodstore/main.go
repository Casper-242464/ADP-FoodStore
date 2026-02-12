package main

import (
	"log"
	"net/http"
	"os"

	"foodstore/config"
	"foodstore/internal/handlers"
	"foodstore/internal/middleware"
	"foodstore/internal/repositories"
	"foodstore/internal/services"
)

func main() {
	cfg := config.GetConfig()
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("failed to prepare upload directory %q: %v", cfg.UploadDir, err)
	}

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
	contactService := services.NewContactService(contactRepo, userRepo)
	userService := services.NewUserService(userRepo)

	ph := handlers.NewProductHandler(productService, userService, cfg.UploadDir)
	oh := handlers.NewOrderHandler(orderService)
	ch := handlers.NewContactHandler(contactService)
	uh := handlers.NewUserHandler(userService)

	http.HandleFunc("/health", handlers.HealthHandler)
	http.Handle("/products", middleware.RequireSeller(userService, http.HandlerFunc(ph.ListProducts)))
	http.HandleFunc("/orders", oh.PlaceOrder)
	http.Handle("/seller/orders", middleware.RequireSellerStrict(userService, http.HandlerFunc(oh.SellerOrders)))
	http.HandleFunc("/contact/messages", ch.ListMessagesForAdmin)
	http.HandleFunc("/contact", ch.HandleContact)

	http.HandleFunc("/api/register", uh.Register)
	http.HandleFunc("/api/login", uh.Login)
	http.HandleFunc("/api/profile", uh.GetProfile)

	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("frontend/styles"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	http.HandleFunc("/ui/products", handlers.ProductsPage)
	http.HandleFunc("/ui/seller/products", handlers.SellerProductsPage)
	http.HandleFunc("/ui/seller/orders", handlers.SellerOrdersPage)
	http.HandleFunc("/ui/orders", handlers.OrdersPage)
	http.HandleFunc("/ui/cart", handlers.CartPage)
	http.HandleFunc("/ui/login", handlers.LoginPage)
	http.HandleFunc("/ui/register", handlers.RegisterPage)
	http.HandleFunc("/ui/profile", handlers.ProfilePage)
	http.HandleFunc("/", handlers.HomePage)

	log.Printf("Server running on %s", cfg.ServerAddress)
	log.Printf("Uploads served from %s", cfg.UploadDir)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, middleware.Logging(http.DefaultServeMux)))
}
