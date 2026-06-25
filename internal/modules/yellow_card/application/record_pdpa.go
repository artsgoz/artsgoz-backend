package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/domain"
)

type RecordPDPA struct {
	repo domain.YellowCardRepository
}

func NewRecordPDPA(repo domain.YellowCardRepository) *RecordPDPA {
	return &RecordPDPA{repo: repo}
}

func (uc *RecordPDPA) Execute(ctx context.Context, userID string, agreed bool) error {
	return uc.repo.RecordPDPAConsent(ctx, userID, agreed)
}
