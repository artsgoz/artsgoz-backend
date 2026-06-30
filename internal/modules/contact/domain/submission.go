package domain

import "time"

type ContactSubmission struct {
	ID        string
	Name      string
	StudentID *string // optional
	Email     string
	Category  string
	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
