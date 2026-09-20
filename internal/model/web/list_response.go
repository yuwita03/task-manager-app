// internal/model/web/list_response.go
package web

type ListResponse struct {
	ID       int    `json:"id"`
	BoardID  int    `json:"board_id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}