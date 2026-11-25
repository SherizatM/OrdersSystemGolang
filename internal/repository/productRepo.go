package repository

import (
	"OrderssSystemGolang/internal/models"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ProductRepo struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func NewProductRepo(pool *pgxpool.Pool, rdb *redis.Client) *ProductRepo {
	return &ProductRepo{pool: pool, rdb: rdb}
}

// GET All products
const productsCacheKey = "products:all"

func (r *ProductRepo) GetProductsList(ctx context.Context) ([]*models.Products, error) {
	cached, err := r.rdb.Get(ctx, productsCacheKey).Result()
	if err == nil {
		log.Println("Hello from redis")
		var products []*models.Products
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			return products, nil
		}
	}

	rows, err := r.pool.Query(ctx, "SELECT * FROM products ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	products, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[models.Products])
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(products); err == nil {
		r.rdb.Set(ctx, productsCacheKey, data, 5*time.Minute)
	}

	return products, nil
}

// POST upload new product
func (r *ProductRepo) CreateProduct(ctx context.Context, product *models.Products) error {
	query := "INSERT INTO products (name, description, price, stock) VALUES ($1, $2, $3, $4) RETURNING id"
	r.rdb.Del(ctx, "products:all")
	return r.pool.QueryRow(ctx, query, product.Name, product.Description, product.Price, product.Stock).Scan(&product.ID)
}

// POST add stock
func (r *ProductRepo) AddStock(ctx context.Context, productId int, quantity int) error {
	query := "UPDATE products SET stock = stock + $1 WHERE id = $2"
	r.rdb.Del(ctx, "products:all")
	_, err := r.pool.Exec(ctx, query, quantity, productId)

	return err
}
