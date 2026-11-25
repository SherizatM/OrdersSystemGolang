package handlers

import (
	"OrderssSystemGolang/internal/repository"
	"encoding/json"
	"net/http"
)

type UserHandler struct {
	repo *repository.UserRepo
}

func CreateNewUserHandler(repo *repository.UserRepo) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) GetAllUsersList(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

}
