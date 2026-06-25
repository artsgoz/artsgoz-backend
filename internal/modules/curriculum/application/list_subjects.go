package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/curriculum/domain"
)

type ListSubjects struct {
	repo domain.CurriculumRepository
}

func NewListSubjects(repo domain.CurriculumRepository) *ListSubjects {
	return &ListSubjects{repo: repo}
}

func (uc *ListSubjects) Execute(ctx context.Context, curriculumID string) (ListSubjectsOutput, error) {
	result, err := uc.repo.FindSubjects(ctx, curriculumID)
	if err != nil {
		return ListSubjectsOutput{}, err
	}

	items := make([]SubjectItem, len(result))
	for i, s := range result {
		items[i] = SubjectItem{
			ID:       s.ID,
			Code:     s.Code,
			NameTh:   s.NameTh,
			NameEn:   s.NameEn,
			Credits:  s.Credits,
			Category: s.Category,
			Semester: s.Semester,
			Group:    s.Group,
			IsCustom: s.IsCustom,
		}
	}

	return ListSubjectsOutput{Subjects: items}, nil
}
