package handlers

import "net/http"

// Only UI pages (HTML). No business logic here.

func HomePage(w http.ResponseWriter, r *http.Request) {
	// serve your mockup main page
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "frontend/pages/index.html")
}

func ProductsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/pages/products.html")
}

func OrdersPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/pages/orders.html")
}

func CartPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/pages/cart.html")
}
