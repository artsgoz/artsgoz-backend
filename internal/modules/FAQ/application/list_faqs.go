package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/FAQ/domain"
)

type ListFAQs struct {
	repo domain.FAQRepository
}

func NewListFAQs(repo domain.FAQRepository) *ListFAQs {
	return &ListFAQs{repo: repo}
}

func (uc *ListFAQs) Execute(ctx context.Context, in ListFAQsInput) ([]FAQItem, error) {
	items, err := uc.repo.List(ctx, in.Tag, in.Q)
	if err != nil {
		return nil, err
	}

	output := make([]FAQItem, len(items))
	for i, f := range items {
		output[i] = FAQItem{
			ID:        f.ID,
			Question:  f.Question,
			Answer:    f.Answer,
			Tags:      f.Tags,
			SortOrder: f.SortOrder,
		}
	}

	return output, nil
}
