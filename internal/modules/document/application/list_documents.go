package application

import (
	"context"
	"math"

	"github.com/artsgoz/artsgoz-backend/internal/modules/document/domain"
)

type ListDocuments struct {
	repo domain.DocumentRepository
}

func NewListDocuments(repo domain.DocumentRepository) *ListDocuments {
	return &ListDocuments{repo: repo}
}

func (uc *ListDocuments) Execute(ctx context.Context, in ListDocumentsInput) (ListDocumentsOutput, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Limit <= 0 {
		in.Limit = 10 // Default 10 ตาม DOCUMENTS_PER_PAGE ใน spec
	}

	result, err := uc.repo.List(ctx, domain.ListFilter{
		Category: in.Category,
		Query:    in.Q,
		Page:     in.Page,
		Limit:    in.Limit,
	})
	if err != nil {
		return ListDocumentsOutput{}, err
	}

	items := make([]DocumentItem, len(result.Items))
	for i, d := range result.Items {
		items[i] = toDocumentItem(d)
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(in.Limit)))

	return ListDocumentsOutput{
		Items: items,
		Pagination: Pagination{
			Page:       in.Page,
			Limit:      in.Limit,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	}, nil
}

func toDocumentItem(d *domain.Document) DocumentItem {
	return DocumentItem{
		ID:       d.ID,
		Name:     d.Name,
		Details:  d.Details,
		Category: d.Category,
		Status:   d.Status,
		FileUrl:  d.FileUrl,
		FileName: d.FileName,
		FileSize: d.FileSize,
	}
}
