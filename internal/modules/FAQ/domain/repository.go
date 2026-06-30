package domain

import "context"

type FAQRepository interface {
	List(ctx context.Context, tag string, query string) ([]*FAQ, error)
	Create(ctx context.Context, faq *FAQ) error
}
