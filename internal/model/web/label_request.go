// internal/model/web/label_request.go
package web

type CreateLabelRequest struct {
	Name  string `json:"name" validate:"required,max=50"`
	Color string `json:"color" validate:"required,len=7"` // format hex, misal #FF5733
}