// internal/model/domain/list.go
package domain

import "time"

type List struct {
	ID        int
	BoardID   int
	Name      string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}