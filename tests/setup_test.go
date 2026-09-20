// tests/setup_test.go
package test

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/app"
	"task-manager-app/internal/config"
)

func setupTestDB() *pgxpool.Pool {
	dsn := "postgres://postgres:postgres@localhost:5432/task_manager_test?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	return pool
}

func setupRouter(db *pgxpool.Pool) http.Handler {
	cfg := config.Config{
		JWTSecret:   "test-secret",
		TokenExpiry: 24 * time.Hour,
	}
	return app.NewRouter(db, cfg)
}

func truncateAll(db *pgxpool.Pool) {
	// urutan CASCADE otomatis handle dependency, tapi tetap urutin biar jelas
	db.Exec(context.Background(), "TRUNCATE task_labels, labels, task_assignees, tasks, lists, board_members, boards, users RESTART IDENTITY CASCADE")
}