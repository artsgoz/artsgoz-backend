package application

import (
	"context"

	"github.com/artsgoz/artsgoz-backend/internal/modules/document/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type UpdateDocument struct {
	repo domain.DocumentRepository
}

func NewUpdateDocument(repo domain.DocumentRepository) *UpdateDocument {
	return &UpdateDocument{repo: repo}
}

func (uc *UpdateDocument) Execute(ctx context.Context, id string, in UpdateDocumentInput) (UpdateDocumentOutput, error) {
	if err := validator.Struct(in); err != nil {
		return UpdateDocumentOutput{}, err
	}

	doc, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return UpdateDocumentOutput{}, err
	}

	doc.Name = in.Name
	doc.Details = in.Details
	doc.Category = in.Category
	doc.Status = in.Status
	doc.FileUrl = in.FileUrl
	doc.FileName = in.FileName
	doc.FileSize = in.FileSize

	if err := uc.repo.Update(ctx, doc); err != nil {
		return UpdateDocumentOutput{}, err
	}

	return UpdateDocumentOutput{ID: id}, nil
}
