package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/domain"
)

type UpdateSubjectCompletion struct {
	repo domain.CreditTrackingRepository
}

func NewUpdateSubjectCompletion(repo domain.CreditTrackingRepository) *UpdateSubjectCompletion {
	return &UpdateSubjectCompletion{repo: repo}
}

func (uc *UpdateSubjectCompletion) Execute(ctx context.Context, userID string, subjectID string, in ToggleCompletionInput) (ToggleCompletionOutput, error) {
	if err := uc.repo.UpdateSubjectCompletion(ctx, userID, subjectID, in.Completed); err != nil {
		return ToggleCompletionOutput{}, err
	}
	return ToggleCompletionOutput{
		ID:        subjectID,
		Completed: in.Completed,
	}, nil
}
