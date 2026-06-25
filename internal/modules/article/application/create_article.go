package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/artsgoz/artsgoz-backend/internal/modules/article/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type CreateArticle struct {
	repo domain.ArticleRepository
}

func NewCreateArticle(repo domain.ArticleRepository) *CreateArticle {
	return &CreateArticle{repo: repo}
}

func (uc *CreateArticle) Execute(ctx context.Context, in CreateArticleInput) (CreateArticleOutput, error) {
	if err := validator.Struct(in); err != nil {
		return CreateArticleOutput{}, err
	}

	id := uuid.NewString()

	art := &domain.Article{
		ID:        id,
		Title:     in.Title,
		Excerpt:   in.Excerpt,
		ImageUrl:  in.ImageUrl,
		Author:    in.Author,
		Category:  in.Category,
		Content:   in.Content,
		IsActive:  true,
	}

	if err := uc.repo.Create(ctx, art); err != nil {
		return CreateArticleOutput{}, err
	}

	return CreateArticleOutput{ID: id}, nil
}
