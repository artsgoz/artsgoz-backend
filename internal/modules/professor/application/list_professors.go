package application

import (
	"context"
	"math"

	"github.com/artsgoz/artsgoz-backend/internal/modules/professor/domain"
)

// ListProfessors is the use case for paginated, filtered professor listing.
type ListProfessors struct {
	repo domain.ProfessorRepository
}

func NewListProfessors(repo domain.ProfessorRepository) *ListProfessors {
	return &ListProfessors{repo: repo}
}

func (uc *ListProfessors) Execute(ctx context.Context, in ListProfessorsInput) (ListProfessorsOutput, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Limit <= 0 {
		in.Limit = 13 // PROFESSORS_PER_PAGE ตาม spec
	}

	result, err := uc.repo.List(ctx, domain.ListFilter{
		Department: in.Department,
		Query:      in.Q,
		Page:       in.Page,
		Limit:      in.Limit,
	})
	if err != nil {
		return ListProfessorsOutput{}, err
	}

	items := make([]ProfessorItem, len(result.Items))
	for i, p := range result.Items {
		items[i] = toProfessorItem(p)
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(in.Limit)))

	return ListProfessorsOutput{
		Items: items,
		Pagination: Pagination{
			Page:       in.Page,
			Limit:      in.Limit,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	}, nil
}

// toProfessorItem maps the domain entity to the API response shape.
func toProfessorItem(p *domain.Professor) ProfessorItem {
	achievements := p.Achievements
	if achievements == nil {
		achievements = []string{}
	}
	qualifications := p.Qualifications
	if qualifications == nil {
		qualifications = []string{}
	}
	courses := p.Courses
	if courses == nil {
		courses = []string{}
	}

	return ProfessorItem{
		ID:             p.ID,
		Name:           p.Name,
		NameEn:         p.NameEn,
		Department:     p.Department,
		Location:       p.Location,
		Achievements:   achievements,
		Qualifications: qualifications,
		Courses:        courses,
	}
}
