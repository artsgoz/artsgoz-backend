package application

import (
	"context"
	"math"

	"github.com/artsgoz/artsgoz-backend/internal/modules/club/domain"
)

type ListClubs struct {
	repo domain.ClubRepository
}

func NewListClubs(repo domain.ClubRepository) *ListClubs {
	return &ListClubs{repo: repo}
}

func (uc *ListClubs) Execute(ctx context.Context, in ListClubsInput) (ListClubsOutput, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Limit <= 0 {
		in.Limit = 12 // Default 12ตาม spec
	}

	result, err := uc.repo.List(ctx, domain.ListFilter{
		Category: in.Category,
		Query:    in.Q,
		Page:     in.Page,
		Limit:    in.Limit,
	})
	if err != nil {
		return ListClubsOutput{}, err
	}

	items := make([]ClubItem, len(result.Items))
	for i, c := range result.Items {
		items[i] = toClubItem(c)
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(in.Limit)))

	return ListClubsOutput{
		Items: items,
		Pagination: Pagination{
			Page:       in.Page,
			Limit:      in.Limit,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	}, nil
}

func toClubItem(c *domain.Club) ClubItem {
	return ClubItem{
		ID:          c.ID,
		Name:        c.Name,
		Category:    c.Category,
		Description: c.Description,
		Instagram:   c.Instagram,
		ImageUrl:    c.ImageUrl,
	}
}
