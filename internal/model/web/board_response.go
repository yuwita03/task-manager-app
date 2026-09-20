// internal/model/web/board_response.go
package web

type BoardResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     int    `json:"owner_id"`
}

type BoardMemberResponse struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}