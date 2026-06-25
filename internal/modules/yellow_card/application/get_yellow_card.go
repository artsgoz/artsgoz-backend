package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/domain"
)

type GetYellowCard struct {
	repo domain.YellowCardRepository
}

func NewGetYellowCard(repo domain.YellowCardRepository) *GetYellowCard {
	return &GetYellowCard{repo: repo}
}

func (uc *GetYellowCard) Execute(ctx context.Context, userID string) (*domain.YellowCard, error) {
	profile, err := uc.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	subjects, err := uc.repo.GetSubjects(ctx, userID)
	if err != nil {
		return nil, err
	}

	gpaTerms, err := uc.repo.GetGPATerms(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.YellowCard{
		PDPAAgreed: profile.PDPAAgreed,
		Profile:    *profile,
		Subjects:   subjects,
		GPAData:    gpaTerms,
	}, nil
}
