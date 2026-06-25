package document

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artsgoz/artsgoz-backend/internal/modules/document/application"
	pgrepo "github.com/artsgoz/artsgoz-backend/internal/modules/document/infrastructure/persistence/postgres"
	"github.com/artsgoz/artsgoz-backend/internal/modules/document/interface/rest"
)

type Module struct {
	Handler *rest.DocumentHandler
}

func New(pool *pgxpool.Pool) *Module {
	repo := pgrepo.NewDocumentRepository(pool)
	listUC := application.NewListDocuments(repo)
	getUC := application.NewGetDocument(repo)
	createUC := application.NewCreateDocument(repo)
	updateUC := application.NewUpdateDocument(repo)
	deleteUC := application.NewDeleteDocument(repo)

	handler := rest.NewDocumentHandler(listUC, getUC, createUC, updateUC, deleteUC, pool)
	return &Module{Handler: handler}
}
