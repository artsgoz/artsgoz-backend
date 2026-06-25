package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/document/domain"
)

type DeleteDocument struct {
	repo domain.DocumentRepository
}

func NewDeleteDocument(repo domain.DocumentRepository) *DeleteDocument {
	return &DeleteDocument{repo: repo}
}

func (uc *DeleteDocument) Execute(ctx context.Context, id string) error {
	_, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.repo.Delete(ctx, id)
}
