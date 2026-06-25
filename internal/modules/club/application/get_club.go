package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/club/domain"
)

type GetClub struct {
	repo domain.ClubRepository
}

func NewGetClub(repo domain.ClubRepository) *GetClub {
	return &GetClub{repo: repo}
}

func (uc *GetClub) Execute(ctx context.Context, id string) (GetClubOutput, error) {
	c, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return GetClubOutput{}, err
	}
	return GetClubOutput{toClubItem(c)}, nil
}
