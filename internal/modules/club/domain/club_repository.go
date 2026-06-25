package domain

import "context"

type ListFilter struct {
	Category string
	Query    string // Search on name and description
	Page       int
	Limit      int
}

type ListResult struct {
	Items []*Club
	Total int
}

type ClubRepository interface {
	List(ctx context.Context, f ListFilter) (ListResult, error)
	FindByID(ctx context.Context, id string) (*Club, error)
	Create(ctx context.Context, club *Club) error
	Update(ctx context.Context, club *Club) error
	Delete(ctx context.Context, id string) error
}
