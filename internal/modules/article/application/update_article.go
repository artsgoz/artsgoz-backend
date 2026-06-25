package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/article/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type UpdateArticle struct {
	repo domain.ArticleRepository
}

func NewUpdateArticle(repo domain.ArticleRepository) *UpdateArticle {
	return &UpdateArticle{repo: repo}
}

func (uc *UpdateArticle) Execute(ctx context.Context, id string, in UpdateArticleInput) (UpdateArticleOutput, error) {
	if err := validator.Struct(in); err != nil {
		return UpdateArticleOutput{}, err
	}

	art, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return UpdateArticleOutput{}, err
	}

	art.Title = in.Title
	art.Excerpt = in.Excerpt
	art.ImageUrl = in.ImageUrl
	art.Author = in.Author
	art.Category = in.Category
	art.Content = in.Content

	if err := uc.repo.Update(ctx, art); err != nil {
		return UpdateArticleOutput{}, err
	}

	return UpdateArticleOutput{ID: id}, nil
}
