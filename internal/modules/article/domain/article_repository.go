package domain

import "context"

type ArticleRepository interface {
	List(ctx context.Context, f ListFilter) (ListResult, error)
	FindByID(ctx context.Context, id string) (*Article, error)
	Create(ctx context.Context, article *Article) error
	Update(ctx context.Context, article *Article) error
	Delete(ctx context.Context, id string) error
}
