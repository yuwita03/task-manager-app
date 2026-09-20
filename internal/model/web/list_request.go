// internal/model/web/list_request.go
package web

type CreateListRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

type UpdateListRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

type ReorderListRequest struct {
	Position int `json:"position" validate:"required,min=0"`
}