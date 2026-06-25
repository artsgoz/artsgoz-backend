package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/domain"
)

type GetSubjects struct {
	repo domain.CreditTrackingRepository
}

func NewGetSubjects(repo domain.CreditTrackingRepository) *GetSubjects {
	return &GetSubjects{repo: repo}
}

func (uc *GetSubjects) Execute(ctx context.Context, userID string) (SubjectsAndProgressOutput, error) {
	sp, err := uc.repo.GetSubjectsAndProgress(ctx, userID)
	if err != nil {
		return SubjectsAndProgressOutput{}, err
	}

	subjects := make([]SubjectItem, len(sp.Subjects))
	for i, s := range sp.Subjects {
		subjects[i] = SubjectItem{
			ID:        s.ID,
			Code:      s.Code,
			NameTh:    s.NameTh,
			NameEn:    s.NameEn,
			Credits:   s.Credits,
			Category:  s.Category,
			Semester:  s.Semester,
			Completed: s.Completed,
			Group:     s.Group,
			IsCustom:  s.IsCustom,
		}
	}

	progress := make([]CategoryProgressItem, len(sp.Progress))
	for i, p := range sp.Progress {
		progress[i] = CategoryProgressItem{
			Category:  p.Category,
			Completed: p.Completed,
			Required:  p.Required,
		}
	}

	return SubjectsAndProgressOutput{
		Profile: ProfileOutput{
			Major:      sp.Profile.Major,
			Minor:      sp.Profile.Minor,
			Curriculum: sp.Profile.Curriculum,
		},
		Subjects: subjects,
		Progress: progress,
	}, nil
}
