package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/domain"
)

type GetProfile struct {
	repo domain.CreditTrackingRepository
}

func NewGetProfile(repo domain.CreditTrackingRepository) *GetProfile {
	return &GetProfile{repo: repo}
}

func (uc *GetProfile) Execute(ctx context.Context, userID string) (ProfileOutput, error) {
	p, err := uc.repo.GetProfile(ctx, userID)
	if err != nil {
		return ProfileOutput{}, err
	}
	return ProfileOutput{
		Major:      p.Major,
		Minor:      p.Minor,
		Curriculum: p.Curriculum,
	}, nil
}
