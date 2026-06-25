package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/artsgoz/artsgoz-backend/internal/modules/club/domain"
)

type CreateClub struct {
	repo domain.ClubRepository
}

func NewCreateClub(repo domain.ClubRepository) *CreateClub {
	return &CreateClub{repo: repo}
}

func (uc *CreateClub) Execute(ctx context.Context, in CreateClubInput) (CreateClubOutput, error) {
	id := uuid.NewString()

	club := &domain.Club{
		ID:          id,
		Name:        in.Name,
		Category:    in.Category,
		Description: in.Description,
		Instagram:   in.Instagram,
		ImageUrl:    in.ImageUrl,
		IsActive:    true,
	}

	if err := uc.repo.Create(ctx, club); err != nil {
		return CreateClubOutput{}, err
	}

	return CreateClubOutput{ID: id}, nil
}
