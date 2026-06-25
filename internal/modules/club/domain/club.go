package domain

import "time"

type Club struct {
	ID          string
	Name        string
	Category    string
	Description string
	Instagram   *string // Nullable
	ImageUrl    *string // Nullable
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
