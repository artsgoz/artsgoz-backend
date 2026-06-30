package domain

import "time"

type FAQ struct {
	ID        string
	Question  string
	Answer    string
	Tags      []string
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
