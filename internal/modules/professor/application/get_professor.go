package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/domain"
)

// GetProfessor is the use case for fetching a single professor by ID.
type GetProfessor struct {
	repo domain.ProfessorRepository
}

func NewGetProfessor(repo domain.ProfessorRepository) *GetProfessor {
	return &GetProfessor{repo: repo}
}

func (uc *GetProfessor) Execute(ctx context.Context, id string) (GetProfessorOutput, error) {
	p, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return GetProfessorOutput{}, err
	}
	return GetProfessorOutput{toProfessorItem(p)}, nil
}
