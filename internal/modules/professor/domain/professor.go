package domain

import "time"

// Professor is the aggregate root for the professor bounded context.
type Professor struct {
	ID             string
	Name           string
	NameEn         *string // Nullable English Name
	Email          string
	Department     string
	Location       string
	Achievements   []string
	Qualifications []string
	Courses        []string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
