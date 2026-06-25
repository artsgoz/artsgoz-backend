package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type UpdateProfile struct {
	repo domain.YellowCardRepository
}

func NewUpdateProfile(repo domain.YellowCardRepository) *UpdateProfile {
	return &UpdateProfile{repo: repo}
}

func (uc *UpdateProfile) Execute(ctx context.Context, userID string, in UpdateProfileInput) error {
	if err := validator.Struct(in); err != nil {
		return err
	}

	profile := &domain.StudentProfile{
		UserID:     userID,
		StudentID:  in.StudentID,
		Name:       in.Name,
		Major:      in.Major,
		Minor:      in.Minor,
		Curriculum: in.Curriculum,
		Advisor:    in.Advisor,
		Address:    in.Address,
		Phone:      in.Phone,
	}

	return uc.repo.UpdateProfile(ctx, profile)
}
