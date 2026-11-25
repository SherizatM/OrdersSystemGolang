package handlers

import (
	"OrderssSystemGolang/internal/models"
	"OrderssSystemGolang/internal/repository"
	"encoding/json"
	"net/http"
)

type OrderHandler struct {
	repo *repository.OrderRepo
}

func NewOrderHandler(repo *repository.OrderRepo) *OrderHandler {
	return &OrderHandler{repo: repo}
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Unauthorazed", http.StatusUnauthorized)
		return
	}

	orders, err := h.repo.GetUserOrders(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}
	if orders == nil {
		orders = []*models.Orders{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

type CreateOrderRequest struct {
	Items []repository.OrderItemInput `json:"items"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "Order must contain at least one item", http.StatusBadRequest)
		return
	}
	for _, item := range req.Items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			http.Error(w, "Invalid product_id or quantity", http.StatusBadRequest)
			return
		}
	}

	order, err := h.repo.CreateOrder(r.Context(), userId, req.Items)
	if err != nil {
		http.Error(w, "Cannot create the order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&order)
}
