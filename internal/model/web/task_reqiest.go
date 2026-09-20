// internal/model/web/task_request.go
package web

type CreateTaskRequest struct {
	Title       string  `json:"title" validate:"required,max=200"`
	Description string  `json:"description" validate:"max=2000"`
	DueDate     *string `json:"due_date"` // format RFC3339, optional
}

type UpdateTaskRequest struct {
	Title       string  `json:"title" validate:"required,max=200"`
	Description string  `json:"description" validate:"max=2000"`
	Status      string  `json:"status" validate:"required,oneof=todo in_progress done"`
	DueDate     *string `json:"due_date"`
}

type MoveTaskRequest struct {
	ListID   int `json:"list_id" validate:"required"`
	Position int `json:"position" validate:"min=0"`
}