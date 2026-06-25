package domain

import "context"

// ListFilter holds optional filters and pagination for the list query.
type ListFilter struct {
	Department string
	Query      string // full-text search on name + department
	Page       int
	Limit      int
}

// ListResult is returned by ProfessorRepository.List.
type ListResult struct {
	Items []*Professor
	Total int
}

// ProfessorRepository is the persistence port for the professor aggregate.
// Implementations live in the infrastructure layer.
type ProfessorRepository interface {
	List(ctx context.Context, f ListFilter) (ListResult, error)
	// FindByID returns the professor or an *apperr.Error of kind not_found.
	FindByID(ctx context.Context, id string) (*Professor, error)
	Create(ctx context.Context, prof *Professor) error
	Update(ctx context.Context, prof *Professor) error
	Delete(ctx context.Context, id string) error
}
