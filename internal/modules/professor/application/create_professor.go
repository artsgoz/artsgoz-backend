package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/domain"
)

type CreateProfessor struct {
	repo domain.ProfessorRepository
}

func NewCreateProfessor(repo domain.ProfessorRepository) *CreateProfessor {
	return &CreateProfessor{repo: repo}
}

func (uc *CreateProfessor) Execute(ctx context.Context, in CreateProfessorInput) (CreateProfessorOutput, error) {
	id := uuid.NewString()

	prof := &domain.Professor{
		ID:             id,
		Name:           in.Name,
		NameEn:         in.NameEn,
		Department:     in.Department,
		Location:       in.Location,
		Achievements:   in.Achievements,
		Qualifications: in.Qualifications,
		Courses:        in.Courses,
		IsActive:       true,
	}

	if err := uc.repo.Create(ctx, prof); err != nil {
		return CreateProfessorOutput{}, err
	}

	return CreateProfessorOutput{ID: id}, nil
}
