package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/artsgoz/artsgoz-backend/internal/modules/contact/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type SubmitContact struct {
	repo domain.ContactSubmissionRepository
}

func NewSubmitContact(repo domain.ContactSubmissionRepository) *SubmitContact {
	return &SubmitContact{repo: repo}
}

func (uc *SubmitContact) Execute(ctx context.Context, in SubmitContactInput) (SubmitContactOutput, error) {
	if err := validator.Struct(in); err != nil {
		return SubmitContactOutput{}, err
	}

	id := uuid.NewString()
	now := time.Now()

	s := &domain.ContactSubmission{
		ID:        id,
		Name:      in.Name,
		StudentID: in.StudentID,
		Email:     in.Email,
		Category:  in.Category,
		Message:   in.Message,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.repo.Create(ctx, s); err != nil {
		return SubmitContactOutput{}, err
	}

	return SubmitContactOutput{ID: id}, nil
}
