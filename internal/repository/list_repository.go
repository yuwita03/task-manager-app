// internal/repository/list_repository.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/domain"
)

type ListRepository interface {
	Save(ctx context.Context, list domain.List) domain.List
	FindById(ctx context.Context, id int) domain.List
	FindByBoardId(ctx context.Context, boardId int) []domain.List
	CountByBoardId(ctx context.Context, boardId int) int
	Update(ctx context.Context, list domain.List) domain.List
	Delete(ctx context.Context, id int)
}

type listRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewListRepository(db *pgxpool.Pool) ListRepository {
	return &listRepositoryImpl{DB: db}
}

func (r *listRepositoryImpl) Save(ctx context.Context, list domain.List) domain.List {
	query := `INSERT INTO lists (board_id, name, position) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := r.DB.QueryRow(ctx, query, list.BoardID, list.Name, list.Position).
		Scan(&list.ID, &list.CreatedAt, &list.UpdatedAt)
	helper.PanicIfError(err)
	return list
}

func (r *listRepositoryImpl) FindById(ctx context.Context, id int) domain.List {
	var list domain.List
	query := `SELECT id, board_id, name, position, created_at, updated_at FROM lists WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).
		Scan(&list.ID, &list.BoardID, &list.Name, &list.Position, &list.CreatedAt, &list.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			panic(exception.NewNotFoundError("list not found"))
		}
		panic(err)
	}
	return list
}

func (r *listRepositoryImpl) FindByBoardId(ctx context.Context, boardId int) []domain.List {
	query := `SELECT id, board_id, name, position, created_at, updated_at FROM lists WHERE board_id = $1 ORDER BY position ASC`
	rows, err := r.DB.Query(ctx, query, boardId)
	helper.PanicIfError(err)
	defer rows.Close()

	var lists []domain.List
	for rows.Next() {
		var list domain.List
		err := rows.Scan(&list.ID, &list.BoardID, &list.Name, &list.Position, &list.CreatedAt, &list.UpdatedAt)
		helper.PanicIfError(err)
		lists = append(lists, list)
	}
	return lists
}

func (r *listRepositoryImpl) CountByBoardId(ctx context.Context, boardId int) int {
	var count int
	query := `SELECT COUNT(*) FROM lists WHERE board_id = $1`
	err := r.DB.QueryRow(ctx, query, boardId).Scan(&count)
	helper.PanicIfError(err)
	return count
}

func (r *listRepositoryImpl) Update(ctx context.Context, list domain.List) domain.List {
	query := `UPDATE lists SET name = $1, position = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.DB.Exec(ctx, query, list.Name, list.Position, list.ID)
	helper.PanicIfError(err)
	return list
}

func (r *listRepositoryImpl) Delete(ctx context.Context, id int) {
	query := `DELETE FROM lists WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	helper.PanicIfError(err)
}