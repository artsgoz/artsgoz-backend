package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/club/domain"
)

type UpdateClub struct {
	repo domain.ClubRepository
}

func NewUpdateClub(repo domain.ClubRepository) *UpdateClub {
	return &UpdateClub{repo: repo}
}

func (uc *UpdateClub) Execute(ctx context.Context, id string, in UpdateClubInput) (UpdateClubOutput, error) {
	club, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return UpdateClubOutput{}, err
	}

	club.Name = in.Name
	club.Category = in.Category
	club.Description = in.Description
	club.Instagram = in.Instagram
	club.ImageUrl = in.ImageUrl

	if err := uc.repo.Update(ctx, club); err != nil {
		return UpdateClubOutput{}, err
	}

	return UpdateClubOutput{ID: id}, nil
}
