// internal/repository/user_repository.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user domain.User) domain.User
	FindByEmail(ctx context.Context, email string) (domain.User, bool)
	FindById(ctx context.Context, id int) domain.User
}

type userRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepositoryImpl{DB: db}
}

func (r *userRepositoryImpl) Save(ctx context.Context, user domain.User) domain.User {
	query := `INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := r.DB.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	helper.PanicIfError(err)
	return user
}

// FindByEmail tetap return (domain.User, bool) BUKAN panic — soalnya "user gak ketemu"
// itu kondisi VALID buat Register (justru dipakai buat cek "email belum kepake"),
// bukan error. Kalau di-panic, Register jadi gak bisa jalan normal.
func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (domain.User, bool) {
	var user domain.User
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1`
	err := r.DB.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return domain.User{}, false
	}
	return user, true
}

func (r *userRepositoryImpl) FindById(ctx context.Context, id int) domain.User {
	var user domain.User
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			panic(exception.NewNotFoundError("user not found"))
		}
		panic(err)
	}
	return user
}