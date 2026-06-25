package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/domain"
)

type GetCurriculum struct {
	repo domain.CurriculumRepository
}

func NewGetCurriculum(repo domain.CurriculumRepository) *GetCurriculum {
	return &GetCurriculum{repo: repo}
}

func (uc *GetCurriculum) Execute(ctx context.Context, id string) (GetCurriculumOutput, error) {
	c, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return GetCurriculumOutput{}, err
	}
	return GetCurriculumOutput{toCurriculumItem(c, true)}, nil
}
