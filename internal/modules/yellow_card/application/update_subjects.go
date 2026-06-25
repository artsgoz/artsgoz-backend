package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/domain"
	"github.com/artsgoz/artsgoz-backend/internal/modules/yellow_card/utils"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type UpdateSubjects struct {
	repo domain.YellowCardRepository
}

func NewUpdateSubjects(repo domain.YellowCardRepository) *UpdateSubjects {
	return &UpdateSubjects{repo: repo}
}

func (uc *UpdateSubjects) Execute(ctx context.Context, userID string, in UpdateSubjectsInput) error {
	if err := validator.Struct(in); err != nil {
		return err
	}

	var subjects []domain.YellowCardSubject
	for i, s := range in.Subjects {
		subjects = append(subjects, domain.YellowCardSubject{
			ID:        s.ID,
			UserID:    userID,
			Code:      s.Code,
			Name:      s.Name,
			Semester:  s.Semester,
			Credits:   s.Credits,
			Grade:     s.Grade,
			Category:  s.Category,
			Group:     s.Group,
			SortOrder: i,
		})
	}

	// Save subjects
	if err := uc.repo.UpdateSubjects(ctx, userID, subjects); err != nil {
		return err
	}

	// Recalculate GPA Data based on the saved subjects
	gpaTerms := utils.CalculateGPA(userID, subjects)
	if err := uc.repo.UpdateGPATerms(ctx, userID, gpaTerms); err != nil {
		return err
	}

	return nil
}
