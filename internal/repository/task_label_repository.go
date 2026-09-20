// internal/repository/task_label_repository.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/domain"
)

type TaskLabelRepository interface {
	Attach(ctx context.Context, taskId int, labelId int)
	Detach(ctx context.Context, taskId int, labelId int)
	FindByTaskId(ctx context.Context, taskId int) []domain.Label
}

type taskLabelRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewTaskLabelRepository(db *pgxpool.Pool) TaskLabelRepository {
	return &taskLabelRepositoryImpl{DB: db}
}

func (r *taskLabelRepositoryImpl) Attach(ctx context.Context, taskId int, labelId int) {
	query := `INSERT INTO task_labels (task_id, label_id) VALUES ($1, $2)`
	_, err := r.DB.Exec(ctx, query, taskId, labelId)
	if err != nil {
		panic(exception.NewConflictError("label already attached to this task"))
	}
}

func (r *taskLabelRepositoryImpl) Detach(ctx context.Context, taskId int, labelId int) {
	query := `DELETE FROM task_labels WHERE task_id = $1 AND label_id = $2`
	_, err := r.DB.Exec(ctx, query, taskId, labelId)
	helper.PanicIfError(err)
}

func (r *taskLabelRepositoryImpl) FindByTaskId(ctx context.Context, taskId int) []domain.Label {
	query := `
		SELECT l.id, l.board_id, l.name, l.color
		FROM task_labels tl
		JOIN labels l ON l.id = tl.label_id
		WHERE tl.task_id = $1`
	rows, err := r.DB.Query(ctx, query, taskId)
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