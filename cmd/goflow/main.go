package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goflow/internal/handler"
	"goflow/internal/metrics"
	"goflow/internal/migrate"
	"goflow/internal/repository"
	"goflow/internal/worker"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://localhost:5432/goflow?sslmode=disable"
	}
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := migrate.Run(connStr, os.DirFS(migrationsDir), "."); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("database ping: %v", err)
	}

	repo := repository.New(pool)
	met := metrics.New()
	cfg := worker.DefaultConfig
	if w := os.Getenv("WORKER_COUNT"); w != "" {
		var n int
		if _, err := fmt.Sscanf(w, "%d", &n); err == nil && n > 0 {
			cfg.WorkerCount = n
		}
	}
	wp := worker.NewPool(repo, met, cfg)
	wp.Run(ctx)

	jh := &handler.JobsHandler{Repo: repo, Metric: met}
	mh := &handler.MetricsHandler{Metric: met}
	hh := &handler.HealthHandler{Check: func() error { return pool.Ping(context.Background()) }}

	r := chi.NewRouter()
	r.Use(handler.RateLimitMiddleware(100))
	r.Get("/health", hh.ServeHTTP)
	r.Post("/jobs", jh.Create)
	r.Get("/jobs/{id}", jh.GetByID)
	r.Get("/metrics", mh.ServeHTTP)

	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("shutdown complete")
}
