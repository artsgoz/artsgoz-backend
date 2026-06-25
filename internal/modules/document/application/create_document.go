package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/artsgoz/artsgoz-backend/internal/modules/document/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type CreateDocument struct {
	repo domain.DocumentRepository
}

func NewCreateDocument(repo domain.DocumentRepository) *CreateDocument {
	return &CreateDocument{repo: repo}
}

func (uc *CreateDocument) Execute(ctx context.Context, in CreateDocumentInput) (CreateDocumentOutput, error) {
	if err := validator.Struct(in); err != nil {
		return CreateDocumentOutput{}, err
	}

	id := uuid.NewString()

	doc := &domain.Document{
		ID:       id,
		Name:     in.Name,
		Details:  in.Details,
		Category: in.Category,
		Status:   in.Status,
		FileUrl:  in.FileUrl,
		FileName: in.FileName,
		FileSize: in.FileSize,
		IsActive: true,
	}

	if err := uc.repo.Create(ctx, doc); err != nil {
		return CreateDocumentOutput{}, err
	}

	return CreateDocumentOutput{ID: id}, nil
}
