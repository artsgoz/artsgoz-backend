package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/artsgoz/artsgoz-backend/internal/modules/FAQ/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type CreateFAQ struct {
	repo domain.FAQRepository
}

func NewCreateFAQ(repo domain.FAQRepository) *CreateFAQ {
	return &CreateFAQ{repo: repo}
}

func (uc *CreateFAQ) Execute(ctx context.Context, in CreateFAQInput) (CreateFAQOutput, error) {
	if err := validator.Struct(in); err != nil {
		return CreateFAQOutput{}, err
	}

	id := uuid.NewString()
	now := time.Now()

	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}

	f := &domain.FAQ{
		ID:        id,
		Question:  in.Question,
		Answer:    in.Answer,
		Tags:      tags,
		SortOrder: in.SortOrder,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.repo.Create(ctx, f); err != nil {
		return CreateFAQOutput{}, err
	}

	return CreateFAQOutput{ID: id}, nil
}
