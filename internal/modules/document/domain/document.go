package domain

import "time"

type Document struct {
	ID        string
	Name      string
	Details   string
	Category  string
	Status    string // failed | pending | neutral | none
	FileUrl   string
	FileName  *string // Nullable
	FileSize  *int64  // Nullable
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
