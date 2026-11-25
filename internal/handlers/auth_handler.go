package handlers

import (
	"OrderssSystemGolang/internal/models"
	"OrderssSystemGolang/internal/repository"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo      *repository.UserRepo
	jwtSecret []byte
}

func NewAuthHandler(repo *repository.UserRepo, jwtSecret string) AuthHandler {
	return AuthHandler{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

type SignUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) SingUpUser(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "name, email and password are required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newUser := models.Users{
		Name:         req.Name,
		Email:        req.Email,
		Role:         "user",
		PasswordHash: string(hash),
	}

	if err := h.repo.CreateUser(r.Context(), &newUser); err != nil {
		http.Error(w, "Failed to create an user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) GenerateJWT(userId int, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) SignInUser(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Cannot sing-in user", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	User, err := h.repo.GetUserByEmail(r.Context(), req.Email)

	if err != nil {
		http.Error(w, "Invalid data", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(User.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid data", http.StatusUnauthorized)
		return
	}

	token, err := h.GenerateJWT(User.ID, User.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})

}
