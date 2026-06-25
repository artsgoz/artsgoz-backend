package domain

import "time"

type Curriculum struct {
	ID           string
	Name         string
	Year         int
	TotalCredits int
	Categories   []CategoryConfig
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CategoryConfig struct {
	Category        string
	RequiredCredits int
	Groups          []string
}

type Subject struct {
	ID           string
	CurriculumID string
	Code         string
	NameTh       string
	NameEn       string
	Credits      float64
	Category     string
	Semester     int
	Group        *string // Nullable
	IsCustom     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
