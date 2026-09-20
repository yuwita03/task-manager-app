// cmd/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"task-manager-app/internal/app"
	"task-manager-app/internal/config"
)

func runMigration(dsn string) {
	m, err := migrate.New("file://migration", dsn)
	if err != nil {
		log.Fatal("gagal load migration:", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("gagal jalanin migration:", err)
	}
}

func setupDatabase(dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("gagal konek database:", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("gagal ping database:", err)
	}
	return pool
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on system env vars")
	}

	// BARU: default ke release mode, kecuali eksplisit di-set "debug"
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	runMigration(cfg.DatabaseURL)

	db := setupDatabase(cfg.DatabaseURL)
	defer db.Close()

	log.Println("database connected, migration up to date")

	r := app.NewRouter(db, cfg)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("server starting on port %s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("server failed:", err)
	}
}