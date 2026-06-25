package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/article/domain"
)

type GetArticle struct {
	repo domain.ArticleRepository
}

func NewGetArticle(repo domain.ArticleRepository) *GetArticle {
	return &GetArticle{repo: repo}
}

func (uc *GetArticle) Execute(ctx context.Context, id string) (GetArticleOutput, error) {
	a, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return GetArticleOutput{}, err
	}

	return GetArticleOutput{
		ID:        a.ID,
		Title:     a.Title,
		Excerpt:   a.Excerpt,
		ImageUrl:  a.ImageUrl,
		Date:      a.CreatedAt,
		Author:    a.Author,
		Category:  a.Category,
		Content:   a.Content,
	}, nil
}
