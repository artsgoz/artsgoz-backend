package domain

import (
	"time"
)

type StudentProfile struct {
	UserID     string
	StudentID  string
	Name       string
	Email      string
	Major      string
	Minor      string
	Curriculum string
	Advisor    string
	Address    string
	Phone      string
	PDPAAgreed bool
}

type YellowCardSubject struct {
	ID        string
	UserID    string
	Code      string
	Name      string
	Semester  string // e.g. "1/2566"
	Credits   string
	Grade     string // A, B+, B, C+, C, D+, D, F, S, U, etc.
	Category  string
	Group     string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GPATerm struct {
	UserID   string
	Semester string // e.g. "1/2566"
	CA       float64
	CG       float64
	GPA      float64
	CAX      float64
	CGX      float64
	GPAX     float64
}

type YellowCard struct {
	PDPAAgreed bool                `json:"pdpaAgreed"`
	Profile    StudentProfile      `json:"profile"`
	Subjects   []YellowCardSubject `json:"subjects"`
	GPAData    []GPATerm           `json:"gpaData"`
}
