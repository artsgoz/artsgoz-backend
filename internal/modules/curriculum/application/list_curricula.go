package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/domain"
)

type ListCurricula struct {
	repo domain.CurriculumRepository
}

func NewListCurricula(repo domain.CurriculumRepository) *ListCurricula {
	return &ListCurricula{repo: repo}
}

func (uc *ListCurricula) Execute(ctx context.Context) (ListCurriculaOutput, error) {
	result, err := uc.repo.List(ctx)
	if err != nil {
		return ListCurriculaOutput{}, err
	}

	items := make([]CurriculumItem, len(result))
	for i, c := range result {
		items[i] = toCurriculumItem(c, false) // Exclude categories for list summary
	}

	return ListCurriculaOutput{Curricula: items}, nil
}

func toCurriculumItem(c *domain.Curriculum, includeCategories bool) CurriculumItem {
	item := CurriculumItem{
		ID:           c.ID,
		Name:         c.Name,
		Year:         c.Year,
		TotalCredits: c.TotalCredits,
	}

	if includeCategories && c.Categories != nil {
		item.Categories = make([]CategoryConfig, len(c.Categories))
		for i, cc := range c.Categories {
			item.Categories[i] = CategoryConfig{
				Category:        cc.Category,
				RequiredCredits: cc.RequiredCredits,
				Groups:          cc.Groups,
			}
		}
	}
	return item
}
