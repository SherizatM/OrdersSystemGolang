package repository

import (
	"OrderssSystemGolang/internal/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type OrderRepo struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func NewOrderRepo(pool *pgxpool.Pool, rdb *redis.Client) *OrderRepo {
	return &OrderRepo{pool: pool, rdb: rdb}
}

type OrderItemInput struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// GET orders
func (r *OrderRepo) GetUserOrders(ctx context.Context, userId int) ([]*models.Orders, error) {
	query := "SELECT * FROM orders WHERE user_id = $1 ORDER BY id DESC"
	rows, err := r.pool.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[models.Orders])
}

// POST order
func (r *OrderRepo) CreateOrder(ctx context.Context, userID int, items []OrderItemInput) (*models.Orders, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	order := &models.Orders{
		UserId: userID,
		Status: "created",
	}

	err = tx.QueryRow(ctx,
		"INSERT INTO orders (user_id, status, total_price) VALUES ($1, $2, 0) RETURNING id, created_at",
		order.UserId, order.Status,
	).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return nil, err
	}

	var totalPrice int

	for _, item := range items {
		var price, stock int

		err = tx.QueryRow(ctx,
			"SELECT price, stock FROM products WHERE id = $1 FOR UPDATE",
			item.ProductID,
		).Scan(&price, &stock)
		if err != nil {
			return nil, err
		}

		if stock < item.Quantity {
			return nil, fmt.Errorf("not enough stock for product %d", item.ProductID)
		}

		_, err = tx.Exec(ctx,
			"UPDATE products SET stock = stock - $1 WHERE id = $2",
			item.Quantity, item.ProductID,
		)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx,
			"INSERT INTO order_items (order_id, product_id, quantity, price_at_purchase) VALUES ($1, $2, $3, $4)",
			order.ID, item.ProductID, item.Quantity, price,
		)
		if err != nil {
			return nil, err
		}

		totalPrice += price * item.Quantity
	}

	_, err = tx.Exec(ctx, "UPDATE orders SET total_price = $1 WHERE id = $2", totalPrice, order.ID)
	if err != nil {
		return nil, err
	}
	order.TotalPrice = totalPrice

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	r.rdb.Del(ctx, "products:all")

	return order, nil
}
