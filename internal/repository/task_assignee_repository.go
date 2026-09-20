// internal/repository/task_assignee_repository.go
package repository

import (
	"context"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/web"
)

type TaskAssigneeRepository interface {
	Assign(ctx context.Context, taskId int, userId int)
	Unassign(ctx context.Context, taskId int, userId int)
	FindByTaskId(ctx context.Context, taskId int) []web.UserResponse
	IsAssigned(ctx context.Context, taskId int, userId int) bool
}

type taskAssigneeRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewTaskAssigneeRepository(db *pgxpool.Pool) TaskAssigneeRepository {
	return &taskAssigneeRepositoryImpl{DB: db}
}

func (r *taskAssigneeRepositoryImpl) Assign(ctx context.Context, taskId int, userId int) {
	query := `INSERT INTO task_assignees (task_id, user_id) VALUES ($1, $2)`
	_, err := r.DB.Exec(ctx, query, taskId, userId)
	if err != nil {
		panic(exception.NewConflictError("user already assigned to this task"))
	}
}

func (r *taskAssigneeRepositoryImpl) Unassign(ctx context.Context, taskId int, userId int) {
	query := `DELETE FROM task_assignees WHERE task_id = $1 AND user_id = $2`
	_, err := r.DB.Exec(ctx, query, taskId, userId)
	helper.PanicIfError(err)
}

func (r *taskAssigneeRepositoryImpl) FindByTaskId(ctx context.Context, taskId int) []web.UserResponse {
	query := `
		SELECT u.id, u.name, u.email, u.created_at
		FROM task_assignees ta
		JOIN users u ON u.id = ta.user_id
		WHERE ta.task_id = $1`
	rows, err := r.DB.Query(ctx, query, taskId)
	helper.PanicIfError(err)
	defer rows.Close()

	var users []web.UserResponse
	for rows.Next() {
		var u web.UserResponse
		var createdAt time.Time
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &createdAt)
		helper.PanicIfError(err)
		u.CreatedAt = createdAt.Format(time.RFC3339)
		users = append(users, u)
	}
	return users
}

func (r *taskAssigneeRepositoryImpl) IsAssigned(ctx context.Context, taskId int, userId int) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM task_assignees WHERE task_id = $1 AND user_id = $2)`
	err := r.DB.QueryRow(ctx, query, taskId, userId).Scan(&exists)
	helper.PanicIfError(err)
	return exists
}