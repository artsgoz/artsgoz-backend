package domain

import "context"

type ListFilter struct {
	Category string
	Query    string // Search on name and details
	Page     int
	Limit    int
}

type ListResult struct {
	Items []*Document
	Total int
}

type DocumentRepository interface {
	List(ctx context.Context, f ListFilter) (ListResult, error)
	FindByID(ctx context.Context, id string) (*Document, error)
	Create(ctx context.Context, doc *Document) error
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, id string) error
}
