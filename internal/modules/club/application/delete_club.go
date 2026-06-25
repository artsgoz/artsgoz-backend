package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/club/domain"
)

type DeleteClub struct {
	repo domain.ClubRepository
}

func NewDeleteClub(repo domain.ClubRepository) *DeleteClub {
	return &DeleteClub{repo: repo}
}

func (uc *DeleteClub) Execute(ctx context.Context, id string) error {
	_, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.repo.Delete(ctx, id)
}
