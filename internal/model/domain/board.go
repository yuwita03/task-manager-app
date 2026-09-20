// internal/model/domain/board.go
package domain

import "time"

type Board struct {
	ID          int
	Name        string
	Description string
	OwnerID     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}