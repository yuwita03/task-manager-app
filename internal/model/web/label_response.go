// internal/model/web/label_response.go
package web

type LabelResponse struct {
	ID      int    `json:"id"`
	BoardID int    `json:"board_id"`
	Name    string `json:"name"`
	Color   string `json:"color"`
}