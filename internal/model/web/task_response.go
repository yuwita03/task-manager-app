// internal/model/web/task_response.go
package web

type TaskResponse struct {
	ID          int             `json:"id"`
	ListID      int             `json:"list_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Position    int             `json:"position"`
	DueDate     string          `json:"due_date,omitempty"`
	CreatedBy   int             `json:"created_by"`
	Assignees   []UserResponse  `json:"assignees"`
	Labels      []LabelResponse `json:"labels"`
}