package handlers

import (
	"html/template"
	"net/http"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
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
	tmpl, err := template.ParseFiles("frontend/pages/orders.html")
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title  string
		Orders []struct {
			ID      int
			UserID  int
			Items   string
			Total   string
			Created string
		}
	}{
		Title: "Your Orders",
		Orders: []struct {
			ID      int
			UserID  int
			Items   string
			Total   string
			Created string
		}{
			{ID: 101, UserID: 7, Items: "2x Apples, 1x Milk", Total: "$12.50", Created: "2026-02-10 22:30"},
			{ID: 102, UserID: 7, Items: "1x Bread, 3x Eggs", Total: "$9.10", Created: "2026-02-09 18:05"},
		},
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "template render error", http.StatusInternalServerError)
		return
	}
}

func CartPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/pages/cart.html")
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/pages/login.html")
}

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/pages/register.html")
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/pages/profile.html")
}
