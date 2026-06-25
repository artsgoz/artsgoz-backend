package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/domain"
)

type UpdateProfile struct {
	repo domain.CreditTrackingRepository
}

func NewUpdateProfile(repo domain.CreditTrackingRepository) *UpdateProfile {
	return &UpdateProfile{repo: repo}
}

func (uc *UpdateProfile) Execute(ctx context.Context, userID string, in ProfileInput) (ProfileOutput, error) {
	p := domain.AcademicProfile{
		Major:      in.Major,
		Minor:      in.Minor,
		Curriculum: in.Curriculum,
	}
	if err := uc.repo.UpdateProfile(ctx, userID, p); err != nil {
		return ProfileOutput{}, err
	}
	return ProfileOutput{
		Major:      p.Major,
		Minor:      p.Minor,
		Curriculum: p.Curriculum,
	}, nil
}
