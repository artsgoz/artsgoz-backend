package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/article/domain"
)

type DeleteArticle struct {
	repo domain.ArticleRepository
}

func NewDeleteArticle(repo domain.ArticleRepository) *DeleteArticle {
	return &DeleteArticle{repo: repo}
}

func (uc *DeleteArticle) Execute(ctx context.Context, id string) error {
	_, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.repo.Delete(ctx, id)
}
