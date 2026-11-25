package handlers

import (
	"OrderssSystemGolang/internal/models"
	"OrderssSystemGolang/internal/repository"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type ProductHandler struct {
	repo *repository.ProductRepo
}

func CreateProductHandler(repo *repository.ProductRepo, rdb *redis.Client) *ProductHandler {
	return &ProductHandler{repo: repo}
}

// GET all products
func (h *ProductHandler) GetProductsList(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.GetProductsList(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// POST new product
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Products

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(product.Name) == "" || product.Price <= 0 || product.Stock < 0 {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.repo.CreateProduct(r.Context(), &product); err != nil {
		http.Error(w, "Falied to create product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

type AddStockReq struct {
	Quantity int `json:"quantity"`
}

// PATCH add stock
func (h *ProductHandler) AddStock(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil || id < 0 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req AddStockReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be greater than 0", http.StatusBadRequest)
		return
	}

	if err := h.repo.AddStock(r.Context(), id, req.Quantity); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "stock updated"})
}
