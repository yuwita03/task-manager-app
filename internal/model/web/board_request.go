// internal/model/web/board_request.go
package web

type CreateBoardRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateBoardRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=1000"`
}

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
}