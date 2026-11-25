package repository

import (
	"OrderssSystemGolang/internal/models"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func CreateNewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Get all usersList
func (r *UserRepo) GetAllUsers(ctx context.Context) ([]models.Users, error) {
	query := "SELECT * FROM users ORDER BY id ASC"
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.Users])
}

// POST (auth New User)
func (r *UserRepo) CreateUser(ctx context.Context, user *models.Users) error {
	query := "INSERT INTO users (name, email, role, password_hash) VALUES ($1, $2, $3, $4) RETURNING id, created_at"
	return r.pool.QueryRow(ctx, query, user.Name, user.Email, user.Role, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
}

// GET (get user by email)
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.Users, error) {
	var User models.Users
	query := "SELECT id, name, email, role, password_hash, created_at FROM users WHERE email = $1"

	err := r.pool.QueryRow(ctx, query, email).Scan(&User.ID, &User.Name, &User.Email, &User.Role, &User.PasswordHash, &User.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &User, nil
}
