// internal/repository/task_repository.go
package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/domain"
)

type TaskRepository interface {
	Save(ctx context.Context, task domain.Task) domain.Task
	FindById(ctx context.Context, id int) domain.Task
	FindByListId(ctx context.Context, listId int) []domain.Task
	CountByListId(ctx context.Context, listId int) int
	Update(ctx context.Context, task domain.Task) domain.Task
	Move(ctx context.Context, taskId int, newListId int, newPosition int) domain.Task
	Delete(ctx context.Context, id int)
}

type taskRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) TaskRepository {
	return &taskRepositoryImpl{DB: db}
}

func (r *taskRepositoryImpl) Save(ctx context.Context, task domain.Task) domain.Task {
	query := `INSERT INTO tasks (list_id, title, description, status, position, due_date, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	err := r.DB.QueryRow(ctx, query, task.ListID, task.Title, task.Description, task.Status, task.Position, task.DueDate, task.CreatedBy).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	helper.PanicIfError(err)
	return task
}

func (r *taskRepositoryImpl) scanTask(row pgx.Row) domain.Task {
	var task domain.Task
	var dueDate *time.Time
	err := row.Scan(&task.ID, &task.ListID, &task.Title, &task.Description, &task.Status,
		&task.Position, &dueDate, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			panic(exception.NewNotFoundError("task not found"))
		}
		panic(err)
	}
	task.DueDate = dueDate
	return task
}

func (r *taskRepositoryImpl) FindById(ctx context.Context, id int) domain.Task {
	query := `SELECT id, list_id, title, description, status, position, due_date, created_by, created_at, updated_at
		FROM tasks WHERE id = $1`
	return r.scanTask(r.DB.QueryRow(ctx, query, id))
}

func (r *taskRepositoryImpl) FindByListId(ctx context.Context, listId int) []domain.Task {
	query := `SELECT id, list_id, title, description, status, position, due_date, created_by, created_at, updated_at
		FROM tasks WHERE list_id = $1 ORDER BY position ASC`
	rows, err := r.DB.Query(ctx, query, listId)
	helper.PanicIfError(err)
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		var dueDate *time.Time
		err := rows.Scan(&task.ID, &task.ListID, &task.Title, &task.Description, &task.Status,
			&task.Position, &dueDate, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt)
		helper.PanicIfError(err)
		task.DueDate = dueDate
		tasks = append(tasks, task)
	}
	return tasks
}

func (r *taskRepositoryImpl) CountByListId(ctx context.Context, listId int) int {
	var count int
	query := `SELECT COUNT(*) FROM tasks WHERE list_id = $1`
	err := r.DB.QueryRow(ctx, query, listId).Scan(&count)
	helper.PanicIfError(err)
	return count
}

func (r *taskRepositoryImpl) Update(ctx context.Context, task domain.Task) domain.Task {
	query := `UPDATE tasks SET title = $1, description = $2, status = $3, due_date = $4, updated_at = NOW() WHERE id = $5`
	_, err := r.DB.Exec(ctx, query, task.Title, task.Description, task.Status, task.DueDate, task.ID)
	helper.PanicIfError(err)
	return task
}

func (r *taskRepositoryImpl) Move(ctx context.Context, taskId int, newListId int, newPosition int) domain.Task {
	query := `UPDATE tasks SET list_id = $1, position = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.DB.Exec(ctx, query, newListId, newPosition, taskId)
	helper.PanicIfError(err)
	return r.FindById(ctx, taskId)
}

func (r *taskRepositoryImpl) Delete(ctx context.Context, id int) {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	helper.PanicIfError(err)
}