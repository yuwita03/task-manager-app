// internal/model/domain/task.go
package domain

import "time"

type Task struct {
	ID          int
	ListID      int
	Title       string
	Description string
	Status      string
	Position    int
	DueDate     *time.Time
	CreatedBy   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}