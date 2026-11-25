package main

import (
	"OrderssSystemGolang/internal/handlers"
	"OrderssSystemGolang/internal/middleware"
	"OrderssSystemGolang/internal/repository"
	"context"
	"fmt"
	"log"
	"net/http"

	"errors"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func runServer(srv *http.Server, rdbCloser interface{ Close() error }, poolCloser interface{ Close() }) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server started on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, stopping server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Forced HTTP server shutdown: %v", err)
	}

	if rdbCloser != nil {
		if err := rdbCloser.Close(); err != nil {
			log.Printf("Redis close error: %v", err)
		}
	}

	if poolCloser != nil {
		poolCloser.Close()
	}

	log.Println("Server stopped cleanly")
}

func main() {
	ctx := context.Background()
	dsn := "postgres://user:password@localhost:5432/orders_db?sslmode=disable"

	pool, err := pgxpool.New(ctx, dsn)

	if err != nil {
		log.Fatalf("Cannot create the pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("DB not working: %v", err)
	}

	fmt.Println("DB was connected")

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	mux := http.NewServeMux()
	secretKey := []byte("Secret-Key")

	// Users
	UserRepo := repository.CreateNewUserRepo(pool)
	UserHandler := handlers.CreateNewUserHandler(UserRepo)
	mux.HandleFunc("GET /users", UserHandler.GetAllUsersList)

	authHandler := handlers.NewAuthHandler(UserRepo, string(secretKey))
	authMid := middleware.AuthMiddleWare(secretKey)

	mux.HandleFunc("POST /auth/sign-up", authHandler.SingUpUser)
	mux.HandleFunc("POST /auth/sign-in", authHandler.SignInUser)

	// products

	ProductsRepo := repository.NewProductRepo(pool, rdb)
	ProductsHandler := handlers.CreateProductHandler(ProductsRepo, rdb)
	mux.HandleFunc("GET /products", ProductsHandler.GetProductsList)
	mux.Handle("POST /products", authMid(http.HandlerFunc(ProductsHandler.CreateProduct)))
	mux.Handle("PATCH /products/{id}/stock", authMid(http.HandlerFunc(ProductsHandler.AddStock)))

	// orders

	OrdersRepo := repository.NewOrderRepo(pool, rdb)
	OrdersHandler := handlers.NewOrderHandler(OrdersRepo)
	mux.Handle("POST /orders", authMid(http.HandlerFunc(OrdersHandler.CreateOrder)))
	mux.Handle("GET /orders", authMid(http.HandlerFunc(OrdersHandler.GetUserOrders)))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	runServer(srv, rdb, pool)

}
