package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/domain"
)

type DeleteCustomSubject struct {
	repo domain.CreditTrackingRepository
}

func NewDeleteCustomSubject(repo domain.CreditTrackingRepository) *DeleteCustomSubject {
	return &DeleteCustomSubject{repo: repo}
}

func (uc *DeleteCustomSubject) Execute(ctx context.Context, userID string, subjectID string) error {
	return uc.repo.DeleteCustomSubject(ctx, userID, subjectID)
}
