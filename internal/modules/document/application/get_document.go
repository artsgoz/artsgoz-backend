package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/document/domain"
)

type GetDocument struct {
	repo domain.DocumentRepository
}

func NewGetDocument(repo domain.DocumentRepository) *GetDocument {
	return &GetDocument{repo: repo}
}

func (uc *GetDocument) Execute(ctx context.Context, id string) (GetDocumentOutput, error) {
	d, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return GetDocumentOutput{}, err
	}
	return GetDocumentOutput{toDocumentItem(d)}, nil
}
