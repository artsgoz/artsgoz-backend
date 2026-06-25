package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/credit_tracking/domain"
)

type AddCustomSubject struct {
	repo domain.CreditTrackingRepository
}

func NewAddCustomSubject(repo domain.CreditTrackingRepository) *AddCustomSubject {
	return &AddCustomSubject{repo: repo}
}

func (uc *AddCustomSubject) Execute(ctx context.Context, userID string, in CustomSubjectInput) (SubjectItem, error) {
	s := domain.Subject{
		Code:     in.Code,
		NameTh:   in.NameTh,
		NameEn:   in.NameEn,
		Credits:  in.Credits,
		Category: in.Category,
		Semester: in.Semester,
	}

	res, err := uc.repo.AddCustomSubject(ctx, userID, s)
	if err != nil {
		return SubjectItem{}, err
	}

	return SubjectItem{
		ID:        res.ID,
		Code:      res.Code,
		NameTh:    res.NameTh,
		NameEn:    res.NameEn,
		Credits:   res.Credits,
		Category:  res.Category,
		Semester:  res.Semester,
		Completed: res.Completed,
		Group:     res.Group,
		IsCustom:  res.IsCustom,
	}, nil
}
