package application

import (
	"context"
	"math"

	"github.com/artsgoz/artsgoz-backend/internal/modules/article/domain"
)

type ListArticles struct {
	repo domain.ArticleRepository
}

func NewListArticles(repo domain.ArticleRepository) *ListArticles {
	return &ListArticles{repo: repo}
}

func (uc *ListArticles) Execute(ctx context.Context, in ListArticlesInput) (ListArticlesOutput, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Limit <= 0 {
		in.Limit = 6 // Default 6 as requested by spec
	}

	result, err := uc.repo.List(ctx, domain.ListFilter{
		Category: in.Category,
		Query:    in.Q,
		Page:     in.Page,
		Limit:    in.Limit,
	})
	if err != nil {
		return ListArticlesOutput{}, err
	}

	items := make([]ArticleItem, len(result.Items))
	for i, a := range result.Items {
		items[i] = ArticleItem{
			ID:       a.ID,
			Title:    a.Title,
			Excerpt:  a.Excerpt,
			ImageUrl: a.ImageUrl,
			Date:     a.CreatedAt,
			Author:   a.Author,
			Category: a.Category,
		}
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(in.Limit)))
	hasNext := in.Page < totalPages
	hasPrev := in.Page > 1

	return ListArticlesOutput{
		Items: items,
		Pagination: Pagination{
			Page:       in.Page,
			Limit:      in.Limit,
			Total:      result.Total,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}, nil
}
