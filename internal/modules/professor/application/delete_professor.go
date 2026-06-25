package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/domain"
)

type DeleteProfessor struct {
	repo domain.ProfessorRepository
}

func NewDeleteProfessor(repo domain.ProfessorRepository) *DeleteProfessor {
	return &DeleteProfessor{repo: repo}
}

func (uc *DeleteProfessor) Execute(ctx context.Context, id string) error {
	_, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.repo.Delete(ctx, id)
}
