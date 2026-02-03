package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var (
	products   = make(map[int]Product)
	productID  = 1
	mu         sync.RWMutex
	orderQueue = make(chan int, 10)
)

func createProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mu.Lock()
	p.ID = productID
	productID++
	products[p.ID] = p
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func listProducts(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	list := make([]Product, 0)
	for _, p := range products {
		list = append(list, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	orderQueue <- 1 

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"order accepted"}`))
}

func orderWorker() {
	for range orderQueue {
		fmt.Println("Processing order...")
		time.Sleep(2 * time.Second)
		fmt.Println("Order completed")
	}
}

func main() {
	go orderWorker() 

	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createProduct(w, r)
		} else if r.Method == http.MethodGet {
			listProducts(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/orders", createOrder)

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
