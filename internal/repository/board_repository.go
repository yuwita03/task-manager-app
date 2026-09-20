// internal/repository/board_repository.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/domain"
)

type BoardRepository interface {
	SaveWithOwner(ctx context.Context, board domain.Board) domain.Board
	FindById(ctx context.Context, id int) domain.Board
	FindByUserId(ctx context.Context, userId int) []domain.Board
	Update(ctx context.Context, board domain.Board) domain.Board
	Delete(ctx context.Context, id int)
}

type boardRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewBoardRepository(db *pgxpool.Pool) BoardRepository {
	return &boardRepositoryImpl{DB: db}
}

func (r *boardRepositoryImpl) SaveWithOwner(ctx context.Context, board domain.Board) domain.Board {
	tx, err := r.DB.Begin(ctx)
	helper.PanicIfError(err)
	defer tx.Rollback(ctx) // no-op kalau udah di-Commit

	query := `INSERT INTO boards (name, description, owner_id) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(ctx, query, board.Name, board.Description, board.OwnerID).
		Scan(&board.ID, &board.CreatedAt, &board.UpdatedAt)
	helper.PanicIfError(err)

	memberQuery := `INSERT INTO board_members (board_id, user_id, role) VALUES ($1, $2, 'owner')`
	_, err = tx.Exec(ctx, memberQuery, board.ID, board.OwnerID)
	helper.PanicIfError(err)

	err = tx.Commit(ctx)
	helper.PanicIfError(err)

	return board
}

func (r *boardRepositoryImpl) FindById(ctx context.Context, id int) domain.Board {
	var board domain.Board
	query := `SELECT id, name, description, owner_id, created_at, updated_at FROM boards WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).
		Scan(&board.ID, &board.Name, &board.Description, &board.OwnerID, &board.CreatedAt, &board.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			panic(exception.NewNotFoundError("board not found"))
		}
		panic(err)
	}
	return board
}

func (r *boardRepositoryImpl) FindByUserId(ctx context.Context, userId int) []domain.Board {
	query := `
		SELECT b.id, b.name, b.description, b.owner_id, b.created_at, b.updated_at
		FROM boards b
		JOIN board_members bm ON bm.board_id = b.id
		WHERE bm.user_id = $1
		ORDER BY b.created_at DESC`
	rows, err := r.DB.Query(ctx, query, userId)
	helper.PanicIfError(err)
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		var board domain.Board
		err := rows.Scan(&board.ID, &board.Name, &board.Description, &board.OwnerID, &board.CreatedAt, &board.UpdatedAt)
		helper.PanicIfError(err)
		boards = append(boards, board)
	}
	return boards
}

func (r *boardRepositoryImpl) Update(ctx context.Context, board domain.Board) domain.Board {
	query := `UPDATE boards SET name = $1, description = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.DB.Exec(ctx, query, board.Name, board.Description, board.ID)
	helper.PanicIfError(err)
	return board
}

func (r *boardRepositoryImpl) Delete(ctx context.Context, id int) {
	query := `DELETE FROM boards WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	helper.PanicIfError(err)
}