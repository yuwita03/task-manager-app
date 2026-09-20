// internal/repository/label_repository.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/domain"
)

type LabelRepository interface {
	Save(ctx context.Context, label domain.Label) domain.Label
	FindById(ctx context.Context, id int) domain.Label
	FindByBoardId(ctx context.Context, boardId int) []domain.Label
	Delete(ctx context.Context, id int)
}

type labelRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewLabelRepository(db *pgxpool.Pool) LabelRepository {
	return &labelRepositoryImpl{DB: db}
}

func (r *labelRepositoryImpl) Save(ctx context.Context, label domain.Label) domain.Label {
	query := `INSERT INTO labels (board_id, name, color) VALUES ($1, $2, $3) RETURNING id`
	err := r.DB.QueryRow(ctx, query, label.BoardID, label.Name, label.Color).Scan(&label.ID)
	helper.PanicIfError(err)
	return label
}

func (r *labelRepositoryImpl) FindById(ctx context.Context, id int) domain.Label {
	var label domain.Label
	query := `SELECT id, board_id, name, color FROM labels WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).Scan(&label.ID, &label.BoardID, &label.Name, &label.Color)
	if err != nil {
		if err == pgx.ErrNoRows {
			panic(exception.NewNotFoundError("label not found"))
		}
		panic(err)
	}
	return label
}

func (r *labelRepositoryImpl) FindByBoardId(ctx context.Context, boardId int) []domain.Label {
	query := `SELECT id, board_id, name, color FROM labels WHERE board_id = $1`
	rows, err := r.DB.Query(ctx, query, boardId)
	helper.PanicIfError(err)
	defer rows.Close()

	var labels []domain.Label
	for rows.Next() {
		var label domain.Label
		err := rows.Scan(&label.ID, &label.BoardID, &label.Name, &label.Color)
		helper.PanicIfError(err)
		labels = append(labels, label)
	}
	return labels
}

func (r *labelRepositoryImpl) Delete(ctx context.Context, id int) {
	query := `DELETE FROM labels WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	helper.PanicIfError(err)
}