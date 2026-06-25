package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/domain"
)

type UpdateProfessor struct {
	repo domain.ProfessorRepository
}

func NewUpdateProfessor(repo domain.ProfessorRepository) *UpdateProfessor {
	return &UpdateProfessor{repo: repo}
}

func (uc *UpdateProfessor) Execute(ctx context.Context, id string, in UpdateProfessorInput) (UpdateProfessorOutput, error) {
	prof, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return UpdateProfessorOutput{}, err
	}

	prof.Name = in.Name
	prof.NameEn = in.NameEn
	prof.Department = in.Department
	prof.Location = in.Location
	prof.Achievements = in.Achievements
	prof.Qualifications = in.Qualifications
	prof.Courses = in.Courses

	if err := uc.repo.Update(ctx, prof); err != nil {
		return UpdateProfessorOutput{}, err
	}

	return UpdateProfessorOutput{ID: id}, nil
}
